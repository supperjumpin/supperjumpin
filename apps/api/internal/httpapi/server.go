     1|     1|     1|package httpapi
     2|     2|     2|
     3|     3|     3|import (
     4|     4|     4|	"encoding/json"
     5|     5|     5|	"errors"
     6|     6|     6|	"io"
     7|     7|     7|	"net/http"
     8|     8|     8|	"strings"
     9|     9|     9|	"time"
    10|    10|    10|)
    11|    11|    11|
    12|    12|    12|type AuthIdentity struct {
    13|    13|    13|	Provider string
    14|    14|    14|	Subject  string
    15|    15|    15|	Email    string
    16|    16|    16|}
    17|    17|    17|
    18|    18|    18|type AuthVerifier interface {
    19|    19|    19|	Verify(token string) (AuthIdentity, bool)
    20|    20|    20|}
    21|    21|    21|
    22|    22|    22|type StaticAuthVerifier map[string]AuthIdentity
    23|    23|    23|
    24|    24|    24|func (v StaticAuthVerifier) Verify(token string) (AuthIdentity, bool) {
    25|    25|    25|	identity, ok := v[token]
    26|    26|    26|	return identity, ok
    27|    27|    27|}
    28|    28|    28|
    29|    29|    29|type ServerConfig struct {
    30|    30|    30|	Auth  AuthVerifier
    31|    31|    31|	Store Store
    32|    32|    32|}
    33|    33|    33|
    34|    34|    34|func NewServer(config ServerConfig) http.Handler {
    35|    35|    35|	mux := http.NewServeMux()
    36|    36|    36|	mux.HandleFunc("GET /v1/me", func(w http.ResponseWriter, r *http.Request) {
    37|    37|    37|		profile, ok := signedInProfile(w, r, config)
    38|    38|    38|		if !ok {
    39|    39|    39|			return
    40|    40|    40|		}
    41|    41|    41|
    42|    42|    42|		writeJSON(w, http.StatusOK, profile)
    43|    43|    43|	})
    44|    44|    44|	mux.HandleFunc("POST /v1/groups", func(w http.ResponseWriter, r *http.Request) {
    45|    45|    45|		profile, ok := signedInProfile(w, r, config)
    46|    46|    46|		if !ok {
    47|    47|    47|			return
    48|    48|    48|		}
    49|    49|    49|
    50|    50|    50|		var request struct {
    51|    51|    51|			Name string `json:"name"`
    52|    52|    52|		}
    53|    53|    53|		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    54|    54|    54|			http.Error(w, "invalid json", http.StatusBadRequest)
    55|    55|    55|			return
    56|    56|    56|		}
    57|    57|    57|		name := strings.TrimSpace(request.Name)
    58|    58|    58|		if name == "" {
    59|    59|    59|			http.Error(w, "Group name is required", http.StatusBadRequest)
    60|    60|    60|			return
    61|    61|    61|		}
    62|    62|    62|
    63|    63|    63|		home, err := config.Store.CreateGroup(r.Context(), profile.Player, name)
    64|    64|    64|		if err != nil {
    65|    65|    65|			http.Error(w, "create Group", http.StatusInternalServerError)
    66|    66|    66|			return
    67|    67|    67|		}
    68|    68|    68|
    69|    69|    69|		writeJSON(w, http.StatusCreated, home)
    70|    70|    70|	})
    71|    71|    71|	mux.HandleFunc("GET /v1/groups", func(w http.ResponseWriter, r *http.Request) {
    72|    72|    72|		profile, ok := signedInProfile(w, r, config)
    73|    73|    73|		if !ok {
    74|    74|    74|			return
    75|    75|    75|		}
    76|    76|    76|
    77|    77|    77|		groups, err := config.Store.ListGroups(r.Context(), profile.Player)
    78|    78|    78|		if err != nil {
    79|    79|    79|			http.Error(w, "list Groups", http.StatusInternalServerError)
    80|    80|    80|			return
    81|    81|    81|		}
    82|    82|    82|
    83|    83|    83|		writeJSON(w, http.StatusOK, groups)
    84|    84|    84|	})
    85|    85|    85|	mux.HandleFunc("GET /v1/groups/{groupID}/home", func(w http.ResponseWriter, r *http.Request) {
    86|    86|    86|		profile, ok := signedInProfile(w, r, config)
    87|    87|    87|		if !ok {
    88|    88|    88|			return
    89|    89|    89|		}
    90|    90|    90|
    91|    91|    91|		home, ok, err := config.Store.GroupHome(r.Context(), profile.Player, r.PathValue("groupID"))
    92|    92|    92|		if err != nil {
    93|    93|    93|			http.Error(w, "get Group home", http.StatusInternalServerError)
    94|    94|    94|			return
    95|    95|    95|		}
    96|    96|    96|		if !ok {
    97|    97|    97|			http.Error(w, "Group Membership required", http.StatusForbidden)
    98|    98|    98|			return
    99|    99|    99|		}
   100|   100|   100|
   101|   101|   101|		writeJSON(w, http.StatusOK, home)
   102|   102|   102|	})
   103|   103|   103|	mux.HandleFunc("POST /v1/groups/{groupID}/invites", func(w http.ResponseWriter, r *http.Request) {
   104|   104|   104|		profile, ok := signedInProfile(w, r, config)
   105|   105|   105|		if !ok {
   106|   106|   106|			return
   107|   107|   107|		}
   108|   108|   108|
   109|   109|   109|		invite, ok, err := config.Store.CreateInvite(r.Context(), profile.Player, r.PathValue("groupID"))
   110|   110|   110|		if err != nil {
   111|   111|   111|			http.Error(w, "create Invite", http.StatusInternalServerError)
   112|   112|   112|			return
   113|   113|   113|		}
   114|   114|   114|		if !ok {
   115|   115|   115|			http.Error(w, "Group Membership required", http.StatusForbidden)
   116|   116|   116|			return
   117|   117|   117|		}
   118|   118|   118|
   119|   119|   119|		writeJSON(w, http.StatusCreated, invite)
   120|   120|   120|	})
   121|   121|   121|	mux.HandleFunc("POST /v1/invites/{token}/accept", func(w http.ResponseWriter, r *http.Request) {
   122|   122|   122|		profile, ok := signedInProfile(w, r, config)
   123|   123|   123|		if !ok {
   124|   124|   124|			return
   125|   125|   125|		}
   126|   126|   126|
   127|   127|   127|		home, status, err := config.Store.AcceptInvite(r.Context(), profile.Player, r.PathValue("token"))
   128|   128|   128|		if err != nil {
   129|   129|   129|			http.Error(w, "accept Invite", http.StatusInternalServerError)
   130|   130|   130|			return
   131|   131|   131|		}
   132|   132|   132|		switch status {
   133|   133|   133|		case InviteInvalid:
   134|   134|   134|			http.Error(w, "Invite cannot be accepted", http.StatusNotFound)
   135|   135|   135|			return
   136|   136|   136|		case InviteUsed:
   137|   137|   137|			http.Error(w, "Invite already used", http.StatusConflict)
   138|   138|   138|			return
   139|   139|   139|		case InviteExpired:
   140|   140|   140|			http.Error(w, "Invite expired", http.StatusGone)
   141|   141|   141|			return
   142|   142|   142|		case InviteMember:
   143|   143|   143|			http.Error(w, "Player already has a Group Membership", http.StatusConflict)
   144|   144|   144|			return
   145|   145|   145|		}
   146|   146|   146|
   147|   147|   147|		writeJSON(w, http.StatusOK, home)
   148|   148|   148|	})
   149|   149|   149|	mux.HandleFunc("POST /v1/groups/{groupID}/seasons", func(w http.ResponseWriter, r *http.Request) {
   150|   150|   150|		profile, ok := signedInProfile(w, r, config)
   151|   151|   151|		if !ok {
   152|   152|   152|			return
   153|   153|   153|		}
   154|   154|   154|
   155|   155|   155|		var body struct {
   156|   156|   156|			SubmissionDeadline string `json:"submissionDeadline"`
   157|   157|   157|			JudgingDeadline    string `json:"judgingDeadline"`
   158|   158|   158|		}
   159|   159|   159|		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
   160|   160|   160|			http.Error(w, "invalid request body", http.StatusBadRequest)
   161|   161|   161|			return
   162|   162|   162|		}
   163|   163|   163|
   164|   164|   164|		submissionDeadline, err := time.Parse(time.RFC3339, body.SubmissionDeadline)
   165|   165|   165|		if err != nil {
   166|   166|   166|			http.Error(w, "submissionDeadline must be ISO 8601 format", http.StatusBadRequest)
   167|   167|   167|			return
   168|   168|   168|		}
   169|   169|   169|		judgingDeadline, err := time.Parse(time.RFC3339, body.JudgingDeadline)
   170|   170|   170|		if err != nil {
   171|   171|   171|			http.Error(w, "judgingDeadline must be ISO 8601 format", http.StatusBadRequest)
   172|   172|   172|			return
   173|   173|   173|		}
   174|   174|   174|
   175|   175|   175|		home, ok, err := config.Store.StartSeason(r.Context(), profile.Player, r.PathValue("groupID"), submissionDeadline, judgingDeadline)
   176|   176|   176|		if errors.Is(err, ErrSeasonAlreadyOpen) {
   177|   177|   177|			http.Error(w, "Group already has an active or closing Season", http.StatusConflict)
   178|   178|   178|			return
   179|   179|   179|		}
   180|   180|   180|		if err != nil {
   181|   181|   181|			http.Error(w, "start Season", http.StatusInternalServerError)
   182|   182|   182|			return
   183|   183|   183|		}
   184|   184|   184|		if !ok {
   185|   185|   185|			http.Error(w, "Group Membership required", http.StatusForbidden)
   186|   186|   186|			return
   187|   187|   187|		}
   188|   188|   188|
   189|   189|   189|		writeJSON(w, http.StatusCreated, home)
   190|   190|   190|	})
   191|   191|   191|	mux.HandleFunc("POST /v1/seasons/{seasonID}/close-submissions", func(w http.ResponseWriter, r *http.Request) {
   192|   192|   192|		profile, ok := signedInProfile(w, r, config)
   193|   193|   193|		if !ok {
   194|   194|   194|			return
   195|   195|   195|		}
   196|   196|   196|
   197|   197|   197|		home, ok, err := config.Store.CloseSeasonSubmissions(r.Context(), profile.Player, r.PathValue("seasonID"))
   198|   198|   198|		if errors.Is(err, ErrSeasonNotFound) {
   199|   199|   199|			http.Error(w, "Season not found", http.StatusNotFound)
   200|   200|   200|			return
   201|   201|   201|		}
   202|   202|   202|		if err != nil {
   203|   203|   203|			http.Error(w, "close Season submissions", http.StatusInternalServerError)
   204|   204|   204|			return
   205|   205|   205|		}
   206|   206|   206|		if !ok {
   207|   207|   207|			http.Error(w, "Season Commissioner or Group Admin required", http.StatusForbidden)
   208|   208|   208|			return
   209|   209|   209|		}
   210|   210|   210|
   211|   211|   211|		writeJSON(w, http.StatusOK, home)
   212|   212|   212|	})
   213|   213|   213|	mux.HandleFunc("POST /v1/seasons/{seasonID}/finalize", func(w http.ResponseWriter, r *http.Request) {
   214|   214|   214|		profile, ok := signedInProfile(w, r, config)
   215|   215|   215|		if !ok {
   216|   216|   216|			return
   217|   217|   217|		}
   218|   218|   218|
   219|   219|   219|		home, ok, err := config.Store.FinalizeSeason(r.Context(), profile.Player, r.PathValue("seasonID"))
   220|   220|   220|		if errors.Is(err, ErrSeasonNotFound) {
   221|   221|   221|			http.Error(w, "Season not found", http.StatusNotFound)
   222|   222|   222|			return
   223|   223|   223|		}
   224|   224|   224|		if err != nil {
   225|   225|   225|			http.Error(w, "finalize Season", http.StatusInternalServerError)
   226|   226|   226|			return
   227|   227|   227|		}
   228|   228|   228|		if !ok {
   229|   229|   229|			http.Error(w, "Season Commissioner or Group Admin required", http.StatusForbidden)
   230|   230|   230|			return
   231|   231|   231|		}
   232|   232|   232|
   233|   233|   233|		writeJSON(w, http.StatusOK, home)
   234|   234|   234|	})
   235|   235|   235|	mux.HandleFunc("GET /v1/seasons/{seasonID}/history", func(w http.ResponseWriter, r *http.Request) {
   236|   236|   236|		profile, ok := signedInProfile(w, r, config)
   237|   237|   237|		if !ok {
   238|   238|   238|			return
   239|   239|   239|		}
   240|   240|   240|
   241|   241|   241|		history, ok, err := config.Store.SeasonHistory(r.Context(), profile.Player, r.PathValue("seasonID"))
   242|   242|   242|		if errors.Is(err, ErrSeasonNotFound) {
   243|   243|   243|			http.Error(w, "Season not found", http.StatusNotFound)
   244|   244|   244|			return
   245|   245|   245|		}
   246|   246|   246|		if err != nil {
   247|   247|   247|			http.Error(w, "get Season history", http.StatusInternalServerError)
   248|   248|   248|			return
   249|   249|   249|		}
   250|   250|   250|		if !ok {
   251|   251|   251|			http.Error(w, "Group Membership required", http.StatusForbidden)
   252|   252|   252|			return
   253|   253|   253|		}
   254|   254|   254|
   255|   255|   255|		writeJSON(w, http.StatusOK, history)
   256|   256|   256|	})
   257|   257|   257|	mux.HandleFunc("POST /v1/groups/{groupID}/ideas", func(w http.ResponseWriter, r *http.Request) {
   258|   258|   258|		profile, ok := signedInProfile(w, r, config)
   259|   259|   259|		if !ok {
   260|   260|   260|			return
   261|   261|   261|		}
   262|   262|   262|
   263|   263|   263|		var request struct {
   264|   264|   264|			Source      string `json:"source"`
   265|   265|   265|			Destination string `json:"destination"`
   266|   266|   266|			Food        string `json:"food"`
   267|   267|   267|		}
   268|   268|   268|		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
   269|   269|   269|			http.Error(w, "invalid json", http.StatusBadRequest)
   270|   270|   270|			return
   271|   271|   271|		}
   272|   272|   272|		source := strings.TrimSpace(request.Source)
   273|   273|   273|		destination := strings.TrimSpace(request.Destination)
   274|   274|   274|		food := strings.TrimSpace(request.Food)
   275|   275|   275|		if source == "" || destination == "" || food == "" {
   276|   276|   276|			http.Error(w, "Source, Destination, and Food are required", http.StatusBadRequest)
   277|   277|   277|			return
   278|   278|   278|		}
   279|   279|   279|
   280|   280|   280|		idea, ok, err := config.Store.CreateIdea(r.Context(), profile.Player, r.PathValue("groupID"), source, destination, food)
   281|   281|   281|		if err != nil {
   282|   282|   282|			http.Error(w, "create Idea", http.StatusInternalServerError)
   283|   283|   283|			return
   284|   284|   284|		}
   285|   285|   285|		if !ok {
   286|   286|   286|			http.Error(w, "Group Membership required", http.StatusForbidden)
   287|   287|   287|			return
   288|   288|   288|		}
   289|   289|   289|
   290|   290|   290|		writeJSON(w, http.StatusCreated, idea)
   291|   291|   291|	})
   292|   292|   292|	mux.HandleFunc("POST /v1/ideas/{ideaID}/planned-jump", func(w http.ResponseWriter, r *http.Request) {
   293|   293|   293|		profile, ok := signedInProfile(w, r, config)
   294|   294|   294|		if !ok {
   295|   295|   295|			return
   296|   296|   296|		}
   297|   297|   297|
   298|   298|   298|		var request struct {
   299|   299|   299|			OffSeason bool `json:"offSeason"`
   300|   300|   300|		}
   301|   301|   301|		if r.Body != nil && r.ContentLength != 0 {
   302|   302|   302|			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
   303|   303|   303|				http.Error(w, "invalid json", http.StatusBadRequest)
   304|   304|   304|				return
   305|   305|   305|			}
   306|   306|   306|		}
   307|   307|   307|		planned, ok, err := config.Store.CreatePlannedStunt(r.Context(), profile.Player, r.PathValue("ideaID"), request.OffSeason)
   308|   308|   308|		if errors.Is(err, ErrStuntNotFound) {
   309|   309|   309|			http.Error(w, "Idea not found", http.StatusNotFound)
   310|   310|   310|			return
   311|   311|   311|		}
   312|   312|   312|		if err != nil {
   313|   313|   313|			http.Error(w, "create Planned Stunt", http.StatusInternalServerError)
   314|   314|   314|			return
   315|   315|   315|		}
   316|   316|   316|		if !ok {
   317|   317|   317|			http.Error(w, "Group Membership required", http.StatusForbidden)
   318|   318|   318|			return
   319|   319|   319|		}
   320|   320|   320|
   321|   321|   321|		writeJSON(w, http.StatusCreated, planned)
   322|   322|   322|	})
   323|   323|   323|	mux.HandleFunc("POST /v1/jumps/{stuntID}/evidence-upload-authorizations", func(w http.ResponseWriter, r *http.Request) {
   324|   324|   324|		profile, ok := signedInProfile(w, r, config)
   325|   325|   325|		if !ok {
   326|   326|   326|			return
   327|   327|   327|		}
   328|   328|   328|
   329|   329|   329|		var request struct {
   330|   330|   330|			ContentType string `json:"contentType"`
   331|   331|   331|		}
   332|   332|   332|		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
   333|   333|   333|			http.Error(w, "invalid json", http.StatusBadRequest)
   334|   334|   334|			return
   335|   335|   335|		}
   336|   336|   336|		contentType := strings.TrimSpace(request.ContentType)
   337|   337|   337|		if contentType == "" {
   338|   338|   338|			http.Error(w, "contentType is required", http.StatusBadRequest)
   339|   339|   339|			return
   340|   340|   340|		}
   341|   341|   341|
   342|   342|   342|		authorization, ok, err := config.Store.AuthorizeEvidenceUpload(r.Context(), profile.Player, r.PathValue("stuntID"), contentType)
   343|   343|   343|		if errors.Is(err, ErrStuntNotFound) {
   344|   344|   344|			http.Error(w, "Planned Stunt not found", http.StatusNotFound)
   345|   345|   345|			return
   346|   346|   346|		}
   347|   347|   347|		if err != nil {
   348|   348|   348|			http.Error(w, "authorize Evidence upload", http.StatusInternalServerError)
   349|   349|   349|			return
   350|   350|   350|		}
   351|   351|   351|		if !ok {
   352|   352|   352|			http.Error(w, "performer required", http.StatusForbidden)
   353|   353|   353|			return
   354|   354|   354|		}
   355|   355|   355|
   356|   356|   356|		writeJSON(w, http.StatusCreated, authorization)
   357|   357|   357|	})
   358|   358|   358|	mux.HandleFunc("POST /v1/jumps/{stuntID}/evidence", func(w http.ResponseWriter, r *http.Request) {
   359|   359|   359|		profile, ok := signedInProfile(w, r, config)
   360|   360|   360|		if !ok {
   361|   361|   361|			return
   362|   362|   362|		}
   363|   363|   363|
   364|   364|   364|		var request struct {
   365|   365|   365|			UploadAuthorizationID string `json:"uploadAuthorizationId"`
   366|   366|   366|			Caption               string `json:"caption"`
   367|   367|   367|		}
   368|   368|   368|		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
   369|   369|   369|			http.Error(w, "invalid json", http.StatusBadRequest)
   370|   370|   370|			return
   371|   371|   371|		}
   372|   372|   372|		uploadAuthorizationID := strings.TrimSpace(request.UploadAuthorizationID)
   373|   373|   373|		caption := strings.TrimSpace(request.Caption)
   374|   374|   374|		if uploadAuthorizationID == "" || caption == "" {
   375|   375|   375|			http.Error(w, "uploadAuthorizationId and caption are required", http.StatusBadRequest)
   376|   376|   376|			return
   377|   377|   377|		}
   378|   378|   378|
   379|   379|   379|		submission, ok, err := config.Store.SubmitEvidence(r.Context(), profile.Player, r.PathValue("stuntID"), uploadAuthorizationID, caption)
   380|   380|   380|		if errors.Is(err, ErrStuntNotFound) {
   381|   381|   381|			http.Error(w, "Planned Stunt not found", http.StatusNotFound)
   382|   382|   382|			return
   383|   383|   383|		}
   384|   384|   384|		if errors.Is(err, ErrEvidenceUploadAuthorizationNotFound) {
   385|   385|   385|			http.Error(w, "Evidence upload authorization not found", http.StatusNotFound)
   386|   386|   386|			return
   387|   387|   387|		}
   388|   388|   388|		if errors.Is(err, ErrSubmissionWindowClosed) {
   389|   389|   389|			http.Error(w, "Submission Window closed", http.StatusConflict)
   390|   390|   390|			return
   391|   391|   391|		}
   392|   392|   392|		if err != nil {
   393|   393|   393|			http.Error(w, "submit Evidence", http.StatusInternalServerError)
   394|   394|   394|			return
   395|   395|   395|		}
   396|   396|   396|		if !ok {
   397|   397|   397|			http.Error(w, "performer required", http.StatusForbidden)
   398|   398|   398|			return
   399|   399|   399|		}
   400|   400|   400|
   401|   401|   401|		writeJSON(w, http.StatusCreated, submission)
   402|   402|   402|	})
   403|   403|   403|	mux.HandleFunc("POST /v1/jumps/{stuntID}/judgment", func(w http.ResponseWriter, r *http.Request) {
   404|   404|   404|		profile, ok := signedInProfile(w, r, config)
   405|   405|   405|		if !ok {
   406|   406|   406|			return
   407|   407|   407|		}
   408|   408|   408|
   409|   409|   409|		var request struct {
   410|   410|   410|			Commitment    *int `json:"commitment"`
   411|   411|   411|			Transgression *int `json:"transgression"`
   412|   412|   412|			Creativity    *int `json:"creativity"`
   413|   413|   413|			Documentation *int `json:"documentation"`
   414|   414|   414|		}
   415|   415|   415|		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
   416|   416|   416|			http.Error(w, "invalid json", http.StatusBadRequest)
   417|   417|   417|			return
   418|   418|   418|		}
   419|   419|   419|		if request.Commitment == nil || request.Transgression == nil || request.Creativity == nil || request.Documentation == nil {
   420|   420|   420|			http.Error(w, "commitment, transgression, creativity, and documentation are required", http.StatusBadRequest)
   421|   421|   421|			return
   422|   422|   422|		}
   423|   423|   423|
   424|   424|   424|		judgment, ok, created, err := config.Store.SubmitJudgment(
   425|   425|   425|			r.Context(),
   426|   426|   426|			profile.Player,
   427|   427|   427|			r.PathValue("stuntID"),
   428|   428|   428|			*request.Commitment,
   429|   429|   429|			*request.Transgression,
   430|   430|   430|			*request.Creativity,
   431|   431|   431|			*request.Documentation,
   432|   432|   432|		)
   433|   433|   433|		if errors.Is(err, ErrStuntNotFound) {
   434|   434|   434|			http.Error(w, "Performed Stunt not found", http.StatusNotFound)
   435|   435|   435|			return
   436|   436|   436|		}
   437|   437|   437|		if errors.Is(err, ErrJudgingWindowClosed) {
   438|   438|   438|			http.Error(w, "Judging Window closed", http.StatusConflict)
   439|   439|   439|			return
   440|   440|   440|		}
   441|   441|   441|		if errors.Is(err, ErrInvalidJudgmentScore) {
   442|   442|   442|			http.Error(w, "Judgment scores must be between 0 and 10", http.StatusBadRequest)
   443|   443|   443|			return
   444|   444|   444|		}
   445|   445|   445|		if err != nil {
   446|   446|   446|			http.Error(w, "submit Judgment", http.StatusInternalServerError)
   447|   447|   447|			return
   448|   448|   448|		}
   449|   449|   449|		if !ok {
   450|   450|   450|			http.Error(w, "Judge required", http.StatusForbidden)
   451|   451|   451|			return
   452|   452|   452|		}
   453|   453|   453|
   454|   454|   454|		status := http.StatusOK
   455|   455|   455|		if created {
   456|   456|   456|			status = http.StatusCreated
   457|   457|   457|		}
   458|   458|   458|		writeJSON(w, status, judgment)
   459|   459|   459|	})
   460|   460|   460|	mux.HandleFunc("POST /v1/jumps/{stuntID}/disputes", func(w http.ResponseWriter, r *http.Request) {
   461|   461|   461|		profile, ok := signedInProfile(w, r, config)
   462|   462|   462|		if !ok {
   463|   463|   463|			return
   464|   464|   464|		}
   465|   465|   465|
   466|   466|   466|		var request struct {
   467|   467|   467|			Concern string `json:"concern"`
   468|   468|   468|			Details string `json:"details"`
   469|   469|   469|		}
   470|   470|   470|		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
   471|   471|   471|			http.Error(w, "invalid json", http.StatusBadRequest)
   472|   472|   472|			return
   473|   473|   473|		}
   474|   474|   474|		concern := strings.TrimSpace(request.Concern)
   475|   475|   475|		details := strings.TrimSpace(request.Details)
   476|   476|   476|		if concern == "" || details == "" {
   477|   477|   477|			http.Error(w, "concern and details are required", http.StatusBadRequest)
   478|   478|   478|			return
   479|   479|   479|		}
   480|   480|   480|
   481|   481|   481|		dispute, ok, err := config.Store.CreateDispute(r.Context(), profile.Player, r.PathValue("stuntID"), concern, details)
   482|   482|   482|		if errors.Is(err, ErrStuntNotFound) {
   483|   483|   483|			http.Error(w, "Visible Stunt not found", http.StatusNotFound)
   484|   484|   484|			return
   485|   485|   485|		}
   486|   486|   486|		if errors.Is(err, ErrInvalidDisputeConcern) {
   487|   487|   487|			http.Error(w, "Dispute concern must be House Rules, Credibility, Source, Destination, Food, duplicate, or other", http.StatusBadRequest)
   488|   488|   488|			return
   489|   489|   489|		}
   490|   490|   490|		if err != nil {
   491|   491|   491|			http.Error(w, "create Dispute", http.StatusInternalServerError)
   492|   492|   492|			return
   493|   493|   493|		}
   494|   494|   494|		if !ok {
   495|   495|   495|			http.Error(w, "Group Membership required", http.StatusForbidden)
   496|   496|   496|			return
   497|   497|   497|		}
   498|   498|   498|
   499|   499|   499|		writeJSON(w, http.StatusCreated, dispute)
   500|   500|   500|	})
   501|