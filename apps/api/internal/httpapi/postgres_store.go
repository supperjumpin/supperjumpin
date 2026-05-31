     1|package httpapi
     2|
     3|import (
     4|	"context"
     5|	"database/sql"
     6|	"errors"
     7|	"fmt"
     8|	"sort"
     9|	"strconv"
    10|	"time"
    11|
    12|	"github.com/jackc/pgx/v5/pgconn"
    13|	_ "github.com/jackc/pgx/v5/stdlib"
    14|)
    15|
    16|type PostgresStore struct {
    17|	db *sql.DB
    18|}
    19|
    20|func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
    21|	db, err := sql.Open("pgx", databaseURL)
    22|	if err != nil {
    23|		return nil, err
    24|	}
    25|	if err := db.PingContext(ctx); err != nil {
    26|		db.Close()
    27|		return nil, err
    28|	}
    29|	return &PostgresStore{db: db}, nil
    30|}
    31|
    32|func (s *PostgresStore) Close() error {
    33|	return s.db.Close()
    34|}
    35|
    36|func (s *PostgresStore) BootstrapIdentity(ctx context.Context, identity AuthIdentity) (MeResponse, error) {
    37|	tx, err := s.db.BeginTx(ctx, nil)
    38|	if err != nil {
    39|		return MeResponse{}, err
    40|	}
    41|	defer tx.Rollback()
    42|
    43|	profile, err := getProfileByAuthIdentity(ctx, tx, identity.Provider, identity.Subject)
    44|	if err == nil {
    45|		return profile, tx.Commit()
    46|	}
    47|	if !errors.Is(err, sql.ErrNoRows) {
    48|		return MeResponse{}, err
    49|	}
    50|
    51|	key := identity.Provider + ":" + identity.Subject
    52|	account := Account{ID: stableID("account", key), Email: identity.Email}
    53|	player := Player{ID: stableID("player", account.ID), DisplayName: displayName(identity.Email)}
    54|	if _, err := tx.ExecContext(ctx, `
    55|INSERT INTO accounts (id, email)
    56|VALUES ($1, $2)
    57|ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email`, account.ID, account.Email); err != nil {
    58|		return MeResponse{}, err
    59|	}
    60|	if _, err := tx.ExecContext(ctx, `
    61|INSERT INTO auth_identities (provider, subject, account_id)
    62|VALUES ($1, $2, $3)
    63|ON CONFLICT (provider, subject) DO NOTHING`, identity.Provider, identity.Subject, account.ID); err != nil {
    64|		return MeResponse{}, err
    65|	}
    66|	if _, err := tx.ExecContext(ctx, `
    67|INSERT INTO players (id, account_id, display_name)
    68|VALUES ($1, $2, $3)
    69|ON CONFLICT (id) DO NOTHING`, player.ID, account.ID, player.DisplayName); err != nil {
    70|		return MeResponse{}, err
    71|	}
    72|
    73|	if err := tx.Commit(); err != nil {
    74|		return MeResponse{}, err
    75|	}
    76|	return MeResponse{Account: account, Player: player}, nil
    77|}
    78|
    79|func (s *PostgresStore) CreateGroup(ctx context.Context, player Player, name string) (GroupHomeResponse, error) {
    80|	tx, err := s.db.BeginTx(ctx, nil)
    81|	if err != nil {
    82|		return GroupHomeResponse{}, err
    83|	}
    84|	defer tx.Rollback()
    85|
    86|	var count int
    87|	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM groups`).Scan(&count); err != nil {
    88|		return GroupHomeResponse{}, err
    89|	}
    90|	group := Group{ID: stableID("group", player.ID+":"+name+":"+strconv.Itoa(count+1)), Name: name}
    91|	membership := GroupMembership{GroupID: group.ID, PlayerID: player.ID, Role: "Group Admin"}
    92|	if _, err := tx.ExecContext(ctx, `INSERT INTO groups (id, name) VALUES ($1, $2)`, group.ID, group.Name); err != nil {
    93|		return GroupHomeResponse{}, err
    94|	}
    95|	if _, err := tx.ExecContext(ctx, `
    96|INSERT INTO group_memberships (group_id, player_id, role)
    97|VALUES ($1, $2, $3)`, membership.GroupID, membership.PlayerID, membership.Role); err != nil {
    98|		return GroupHomeResponse{}, err
    99|	}
   100|	if err := tx.Commit(); err != nil {
   101|		return GroupHomeResponse{}, err
   102|	}
   103|	return groupHome(group, membership, nil, []PerformedStuntView{}, []StandingEntry{}), nil
   104|}
   105|
   106|func (s *PostgresStore) GroupHome(ctx context.Context, player Player, groupID string) (GroupHomeResponse, bool, error) {
   107|	if err := s.ensureSeasonStatusesForGroup(ctx, groupID); err != nil {
   108|		return GroupHomeResponse{}, false, err
   109|	}
   110|	var group Group
   111|	var membership GroupMembership
   112|	if err := s.db.QueryRowContext(ctx, `
   113|SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
   114|FROM group_memberships
   115|JOIN groups ON groups.id = group_memberships.group_id
   116|WHERE group_memberships.group_id = $1 AND group_memberships.player_id = $2`, groupID, player.ID).Scan(
   117|		&group.ID,
   118|		&group.Name,
   119|		&membership.PlayerID,
   120|		&membership.Role,
   121|	); err != nil {
   122|		if errors.Is(err, sql.ErrNoRows) {
   123|			return GroupHomeResponse{}, false, nil
   124|		}
   125|		return GroupHomeResponse{}, false, err
   126|	}
   127|	membership.GroupID = group.ID
   128|	season, err := s.currentSeasonForGroup(ctx, group.ID)
   129|	if err != nil {
   130|		return GroupHomeResponse{}, false, err
   131|	}
   132|	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, group.ID)
   133|	if err != nil {
   134|		return GroupHomeResponse{}, false, err
   135|	}
   136|	standings, err := s.standingsForGroup(ctx, group.ID)
   137|	if err != nil {
   138|		return GroupHomeResponse{}, false, err
   139|	}
   140|	return groupHome(group, membership, season, recentStunts, standings), true, nil
   141|}
   142|
   143|func (s *PostgresStore) ListGroups(ctx context.Context, player Player) (ListGroupsResponse, error) {
   144|	rows, err := s.db.QueryContext(ctx, `
   145|SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
   146|FROM group_memberships
   147|JOIN groups ON groups.id = group_memberships.group_id
   148|WHERE group_memberships.player_id = $1
   149|ORDER BY groups.name`, player.ID)
   150|	if err != nil {
   151|		return ListGroupsResponse{}, err
   152|	}
   153|	defer rows.Close()
   154|
   155|	memberships := []GroupMembershipSummary{}
   156|	for rows.Next() {
   157|		var group Group
   158|		var membership GroupMembership
   159|		if err := rows.Scan(&group.ID, &group.Name, &membership.PlayerID, &membership.Role); err != nil {
   160|			return ListGroupsResponse{}, err
   161|		}
   162|		membership.GroupID = group.ID
   163|		memberships = append(memberships, GroupMembershipSummary{Group: group, Membership: membership})
   164|	}
   165|	if err := rows.Err(); err != nil {
   166|		return ListGroupsResponse{}, err
   167|	}
   168|	sort.Slice(memberships, func(i, j int) bool {
   169|		return memberships[i].Group.Name < memberships[j].Group.Name
   170|	})
   171|	return ListGroupsResponse{Memberships: memberships}, nil
   172|}
   173|
   174|func (s *PostgresStore) CreateInvite(ctx context.Context, player Player, groupID string) (Invite, bool, error) {
   175|	tx, err := s.db.BeginTx(ctx, nil)
   176|	if err != nil {
   177|		return Invite{}, false, err
   178|	}
   179|	defer tx.Rollback()
   180|
   181|	var existing string
   182|	if err := tx.QueryRowContext(ctx, `
   183|SELECT player_id
   184|FROM group_memberships
   185|WHERE group_id = $1 AND player_id = $2`, groupID, player.ID).Scan(&existing); err != nil {
   186|		if errors.Is(err, sql.ErrNoRows) {
   187|			return Invite{}, false, nil
   188|		}
   189|		return Invite{}, false, err
   190|	}
   191|
   192|	var invite Invite
   193|	for attempts := 0; attempts < 3; attempts++ {
   194|		id, err := randomToken("invite")
   195|		if err != nil {
   196|			return Invite{}, false, err
   197|		}
   198|		token, err := randomToken("invite_token")
   199|		if err != nil {
   200|			return Invite{}, false, err
   201|		}
   202|		invite = Invite{
   203|			ID:        id,
   204|			GroupID:   groupID,
   205|			Token:     token,
   206|			CreatedBy: player.ID,
   207|			ExpiresAt: time.Now().Add(7 * 24 * time.Hour).UTC(),
   208|		}
   209|		result, err := tx.ExecContext(ctx, `
   210|INSERT INTO invites (id, group_id, token, created_by_player_id, expires_at)
   211|VALUES ($1, $2, $3, $4, $5)
   212|ON CONFLICT DO NOTHING`, invite.ID, invite.GroupID, invite.Token, invite.CreatedBy, invite.ExpiresAt)
   213|		if err != nil {
   214|			return Invite{}, false, err
   215|		}
   216|		rows, err := result.RowsAffected()
   217|		if err != nil {
   218|			return Invite{}, false, err
   219|		}
   220|		if rows == 1 {
   221|			break
   222|		}
   223|		if attempts == 2 {
   224|			return Invite{}, false, fmt.Errorf("create unique Invite after retries")
   225|		}
   226|	}
   227|	if err := tx.Commit(); err != nil {
   228|		return Invite{}, false, err
   229|	}
   230|	return invite, true, nil
   231|}
   232|
   233|func (s *PostgresStore) AcceptInvite(ctx context.Context, player Player, token string) (GroupHomeResponse, InviteAcceptStatus, error) {
   234|	tx, err := s.db.BeginTx(ctx, nil)
   235|	if err != nil {
   236|		return GroupHomeResponse{}, InviteInvalid, err
   237|	}
   238|	defer tx.Rollback()
   239|
   240|	var invite Invite
   241|	var usedBy sql.NullString
   242|	if err := tx.QueryRowContext(ctx, `
   243|SELECT id, group_id, token, created_by_player_id, used_by_player_id, expires_at
   244|FROM invites
   245|WHERE token = $1`, token).Scan(&invite.ID, &invite.GroupID, &invite.Token, &invite.CreatedBy, &usedBy, &invite.ExpiresAt); err != nil {
   246|		if errors.Is(err, sql.ErrNoRows) {
   247|			return GroupHomeResponse{}, InviteInvalid, nil
   248|		}
   249|		return GroupHomeResponse{}, InviteInvalid, err
   250|	}
   251|	if usedBy.Valid {
   252|		return GroupHomeResponse{}, InviteUsed, nil
   253|	}
   254|	if time.Now().After(invite.ExpiresAt) {
   255|		return GroupHomeResponse{}, InviteExpired, nil
   256|	}
   257|	var existingRole string
   258|	if err := tx.QueryRowContext(ctx, `
   259|SELECT role
   260|FROM group_memberships
   261|WHERE group_id = $1 AND player_id = $2`, invite.GroupID, player.ID).Scan(&existingRole); err != nil && !errors.Is(err, sql.ErrNoRows) {
   262|		return GroupHomeResponse{}, InviteInvalid, err
   263|	} else if err == nil {
   264|		return GroupHomeResponse{}, InviteMember, nil
   265|	}
   266|
   267|	var group Group
   268|	if err := tx.QueryRowContext(ctx, `SELECT id, name FROM groups WHERE id = $1`, invite.GroupID).Scan(&group.ID, &group.Name); err != nil {
   269|		if errors.Is(err, sql.ErrNoRows) {
   270|			return GroupHomeResponse{}, InviteInvalid, nil
   271|		}
   272|		return GroupHomeResponse{}, InviteInvalid, err
   273|	}
   274|	if err := tx.QueryRowContext(ctx, `
   275|UPDATE invites
   276|SET used_by_player_id = $1
   277|WHERE token = $2 AND used_by_player_id IS NULL AND expires_at > now()
   278|RETURNING id`, player.ID, token).Scan(&invite.ID); err != nil {
   279|		if !errors.Is(err, sql.ErrNoRows) {
   280|			return GroupHomeResponse{}, InviteInvalid, err
   281|		}
   282|		if status, err := s.inviteStatus(ctx, tx, token); err != nil || status != InviteInvalid {
   283|			return GroupHomeResponse{}, status, err
   284|		}
   285|		return GroupHomeResponse{}, InviteInvalid, nil
   286|	}
   287|	membership := GroupMembership{GroupID: invite.GroupID, PlayerID: player.ID, Role: "Player"}
   288|	if _, err := tx.ExecContext(ctx, `
   289|INSERT INTO group_memberships (group_id, player_id, role)
   290|VALUES ($1, $2, $3)
   291|ON CONFLICT (group_id, player_id) DO NOTHING`, membership.GroupID, membership.PlayerID, membership.Role); err != nil {
   292|		return GroupHomeResponse{}, InviteInvalid, err
   293|	}
   294|	if err := tx.QueryRowContext(ctx, `
   295|SELECT role
   296|FROM group_memberships
   297|WHERE group_id = $1 AND player_id = $2`, invite.GroupID, player.ID).Scan(&membership.Role); err != nil {
   298|		return GroupHomeResponse{}, InviteInvalid, err
   299|	}
   300|	if err := tx.Commit(); err != nil {
   301|		return GroupHomeResponse{}, InviteInvalid, err
   302|	}
   303|	if err := s.ensureSeasonStatusesForGroup(ctx, group.ID); err != nil {
   304|		return GroupHomeResponse{}, InviteInvalid, err
   305|	}
   306|	season, err := s.currentSeasonForGroup(ctx, group.ID)
   307|	if err != nil {
   308|		return GroupHomeResponse{}, InviteInvalid, err
   309|	}
   310|	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, invite.GroupID)
   311|	if err != nil {
   312|		return GroupHomeResponse{}, InviteInvalid, err
   313|	}
   314|	standings, err := s.standingsForGroup(ctx, group.ID)
   315|	if err != nil {
   316|		return GroupHomeResponse{}, InviteInvalid, err
   317|	}
   318|	return groupHome(group, membership, season, recentStunts, standings), InviteAccepted, nil
   319|}
   320|
   321|func (s *PostgresStore) StartSeason(ctx context.Context, player Player, groupID string, submissionDeadline time.Time, judgingDeadline time.Time) (GroupHomeResponse, bool, error) {
   322|	tx, err := s.db.BeginTx(ctx, nil)
   323|	if err != nil {
   324|		return GroupHomeResponse{}, false, err
   325|	}
   326|	defer tx.Rollback()
   327|
   328|	var group Group
   329|	var membership GroupMembership
   330|	if err := tx.QueryRowContext(ctx, `
   331|SELECT groups.id, groups.name, group_memberships.player_id, group_memberships.role
   332|FROM group_memberships
   333|JOIN groups ON groups.id = group_memberships.group_id
   334|WHERE group_memberships.group_id = $1 AND group_memberships.player_id = $2`, groupID, player.ID).Scan(
   335|		&group.ID,
   336|		&group.Name,
   337|		&membership.PlayerID,
   338|		&membership.Role,
   339|	); err != nil {
   340|		if errors.Is(err, sql.ErrNoRows) {
   341|			return GroupHomeResponse{}, false, nil
   342|		}
   343|		return GroupHomeResponse{}, false, err
   344|	}
   345|	membership.GroupID = group.ID
   346|	var openCount int
   347|	if err := tx.QueryRowContext(ctx, `
   348|SELECT count(*)
   349|FROM seasons
   350|WHERE group_id = $1 AND status IN ('Active', 'Judging Grace Period')`, groupID).Scan(&openCount); err != nil {
   351|		return GroupHomeResponse{}, false, err
   352|	}
   353|	if openCount > 0 {
   354|		return GroupHomeResponse{}, true, ErrSeasonAlreadyOpen
   355|	}
   356|
   357|	var count int
   358|	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM seasons`).Scan(&count); err != nil {
   359|		return GroupHomeResponse{}, false, err
   360|	}
   361|	season := Season{
   362|		ID:                   stableID("season", groupID+":"+strconv.Itoa(count+1)),
   363|		GroupID:              groupID,
   364|		CommissionerPlayerID: player.ID,
   365|		Status:               "Active",
   366|		SubmissionDeadline:   submissionDeadline.UTC(),
   367|		JudgingDeadline:      judgingDeadline.UTC(),
   368|	}
   369|	if _, err := tx.ExecContext(ctx, `
   370|INSERT INTO seasons (id, group_id, commissioner_player_id, status, submission_deadline, judging_deadline)
   371|VALUES ($1, $2, $3, $4, $5, $6)`, season.ID, season.GroupID, season.CommissionerPlayerID, season.Status, season.SubmissionDeadline, season.JudgingDeadline); err != nil {
   372|		if isSeasonOpenConflict(err) {
   373|			return GroupHomeResponse{}, true, ErrSeasonAlreadyOpen
   374|		}
   375|		return GroupHomeResponse{}, false, err
   376|	}
   377|	if err := tx.Commit(); err != nil {
   378|		return GroupHomeResponse{}, false, err
   379|	}
   380|	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, groupID)
   381|	if err != nil {
   382|		return GroupHomeResponse{}, false, err
   383|	}
   384|	return groupHome(group, membership, &season, recentStunts, []StandingEntry{}), true, nil
   385|}
   386|
   387|func (s *PostgresStore) CloseSeasonSubmissions(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
   388|	if err := s.ensureSeasonStatusesForSeason(ctx, seasonID); err != nil {
   389|		return GroupHomeResponse{}, false, err
   390|	}
   391|
   392|	tx, err := s.db.BeginTx(ctx, nil)
   393|	if err != nil {
   394|		return GroupHomeResponse{}, false, err
   395|	}
   396|	defer tx.Rollback()
   397|
   398|	var group Group
   399|	var membership GroupMembership
   400|	var season Season
   401|	if err := tx.QueryRowContext(ctx, `
   402|SELECT seasons.id, seasons.group_id, seasons.commissioner_player_id, seasons.status, seasons.submission_deadline, seasons.judging_deadline,
   403|       groups.name, group_memberships.player_id, group_memberships.role
   404|FROM seasons
   405|JOIN groups ON groups.id = seasons.group_id
   406|JOIN group_memberships ON group_memberships.group_id = seasons.group_id
   407|WHERE seasons.id = $1 AND group_memberships.player_id = $2`, seasonID, player.ID).Scan(
   408|		&season.ID,
   409|		&season.GroupID,
   410|		&season.CommissionerPlayerID,
   411|		&season.Status,
   412|		&season.SubmissionDeadline,
   413|		&season.JudgingDeadline,
   414|		&group.Name,
   415|		&membership.PlayerID,
   416|		&membership.Role,
   417|	); err != nil {
   418|		if errors.Is(err, sql.ErrNoRows) {
   419|			var exists bool
   420|			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM seasons WHERE id = $1)`, seasonID).Scan(&exists); err != nil {
   421|				return GroupHomeResponse{}, false, err
   422|			}
   423|			if !exists {
   424|				return GroupHomeResponse{}, false, ErrSeasonNotFound
   425|			}
   426|			return GroupHomeResponse{}, false, nil
   427|		}
   428|		return GroupHomeResponse{}, false, err
   429|	}
   430|	group.ID = season.GroupID
   431|	membership.GroupID = season.GroupID
   432|	if player.ID != season.CommissionerPlayerID && membership.Role != "Group Admin" {
   433|		return GroupHomeResponse{}, false, nil
   434|	}
   435|	if season.Status == "Active" {
   436|		fromStatus := season.Status
   437|		if _, err := tx.ExecContext(ctx, `UPDATE seasons SET status = 'Judging Grace Period' WHERE id = $1`, season.ID); err != nil {
   438|			return GroupHomeResponse{}, false, err
   439|		}
   440|		season.Status = "Judging Grace Period"
   441|		if err := insertSeasonHistoryEntry(ctx, tx, season.ID, "Submissions Closed", player.ID, membership.Role, player.ID != season.CommissionerPlayerID, fromStatus, season.Status); err != nil {
   442|			return GroupHomeResponse{}, false, err
   443|		}
   444|	}
   445|	if err := tx.Commit(); err != nil {
   446|		return GroupHomeResponse{}, false, err
   447|	}
   448|
   449|	recentStunts, err := recentPerformedStuntsForGroupQuery(ctx, s.db, season.GroupID)
   450|	if err != nil {
   451|		return GroupHomeResponse{}, false, err
   452|	}
   453|	standings, err := s.standingsForGroup(ctx, season.GroupID)
   454|	if err != nil {
   455|		return GroupHomeResponse{}, false, err
   456|	}
   457|	currentSeason, err := s.currentSeasonForGroup(ctx, season.GroupID)
   458|	if err != nil {
   459|		return GroupHomeResponse{}, false, err
   460|	}
   461|	return groupHome(group, membership, currentSeason, recentStunts, standings), true, nil
   462|}
   463|
   464|func (s *PostgresStore) FinalizeSeason(ctx context.Context, player Player, seasonID string) (GroupHomeResponse, bool, error) {
   465|	if err := s.ensureSeasonStatusesForSeason(ctx, seasonID); err != nil {
   466|		return GroupHomeResponse{}, false, err
   467|	}
   468|
   469|	tx, err := s.db.BeginTx(ctx, nil)
   470|	if err != nil {
   471|		return GroupHomeResponse{}, false, err
   472|	}
   473|	defer tx.Rollback()
   474|
   475|	var group Group
   476|	var membership GroupMembership
   477|	var season Season
   478|	if err := tx.QueryRowContext(ctx, `
   479|SELECT seasons.id, seasons.group_id, seasons.commissioner_player_id, seasons.status, seasons.submission_deadline, seasons.judging_deadline,
   480|       groups.name, group_memberships.player_id, group_memberships.role
   481|FROM seasons
   482|JOIN groups ON groups.id = seasons.group_id
   483|JOIN group_memberships ON group_memberships.group_id = seasons.group_id
   484|WHERE seasons.id = $1 AND group_memberships.player_id = $2`, seasonID, player.ID).Scan(
   485|		&season.ID,
   486|		&season.GroupID,
   487|		&season.CommissionerPlayerID,
   488|		&season.Status,
   489|		&season.SubmissionDeadline,
   490|		&season.JudgingDeadline,
   491|		&group.Name,
   492|		&membership.PlayerID,
   493|		&membership.Role,
   494|	); err != nil {
   495|		if errors.Is(err, sql.ErrNoRows) {
   496|			var exists bool
   497|			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM seasons WHERE id = $1)`, seasonID).Scan(&exists); err != nil {
   498|				return GroupHomeResponse{}, false, err
   499|			}
   500|			if !exists {
   501|