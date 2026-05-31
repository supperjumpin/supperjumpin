     1|import type {
     2|  EvidenceSubmission,
     3|  EvidenceUploadAuthorization,
     4|  GroupHomeResponse,
     5|  Invite,
     6|  Judgment,
     7|  ListGroupsResponse,
     8|  MeResponse,
     9|  PerformedStuntView,
    10|  Stunt,
    11|} from "./generated";
    12|
    13|export type {
    14|  Account,
    15|  Group,
    16|  GroupHomeResponse,
    17|  GroupMembership,
    18|  GroupMembershipSummary,
    19|  Evidence,
    20|  EvidenceSubmission,
    21|  EvidenceUploadAuthorization,
    22|  Invite,
    23|  Judgment,
    24|  ListGroupsResponse,
    25|  MeResponse,
    26|  PerformedStuntView,
    27|  Player,
    28|  Season,
    29|  Stunt,
    30|} from "./generated";
    31|
    32|export function getMe(args: {
    33|  baseUrl: string;
    34|  accessToken: string;
    35|  fetchImpl?: typeof fetch;
    36|}): Promise<MeResponse>;
    37|
    38|export function createGroup(args: {
    39|  baseUrl: string;
    40|  accessToken: string;
    41|  name: string;
    42|  fetchImpl?: typeof fetch;
    43|}): Promise<GroupHomeResponse>;
    44|
    45|export function listGroups(args: {
    46|  baseUrl: string;
    47|  accessToken: string;
    48|  fetchImpl?: typeof fetch;
    49|}): Promise<ListGroupsResponse>;
    50|
    51|export function getGroupHome(args: {
    52|  baseUrl: string;
    53|  accessToken: string;
    54|  groupId: string;
    55|  fetchImpl?: typeof fetch;
    56|}): Promise<GroupHomeResponse>;
    57|
    58|export function createInvite(args: {
    59|  baseUrl: string;
    60|  accessToken: string;
    61|  groupId: string;
    62|  fetchImpl?: typeof fetch;
    63|}): Promise<Invite>;
    64|
    65|export function acceptInvite(args: {
    66|  baseUrl: string;
    67|  accessToken: string;
    68|  token: string;
    69|  fetchImpl?: typeof fetch;
    70|}): Promise<GroupHomeResponse>;
    71|
    72|export function startSeason(args: {
    73|  baseUrl: string;
    74|  accessToken: string;
    75|  groupId: string;
    76|  fetchImpl?: typeof fetch;
    77|}): Promise<GroupHomeResponse>;
    78|
    79|export function createIdea(args: {
    80|  baseUrl: string;
    81|  accessToken: string;
    82|  groupId: string;
    83|  source: string;
    84|  destination: string;
    85|  food: string;
    86|  fetchImpl?: typeof fetch;
    87|}): Promise<Stunt>;
    88|
    89|export function createPlannedStunt(args: {
    90|  baseUrl: string;
    91|  accessToken: string;
    92|  ideaId: string;
    93|  offSeason?: boolean;
    94|  fetchImpl?: typeof fetch;
    95|}): Promise<Stunt>;
    96|
    97|export function authorizeEvidenceUpload(args: {
    98|  baseUrl: string;
    99|  accessToken: string;
   100|  stuntId: string;
   101|  contentType: string;
   102|  fetchImpl?: typeof fetch;
   103|}): Promise<EvidenceUploadAuthorization>;
   104|
   105|export function submitEvidence(args: {
   106|  baseUrl: string;
   107|  accessToken: string;
   108|  stuntId: string;
   109|  uploadAuthorizationId: string;
   110|  caption: string;
   111|  fetchImpl?: typeof fetch;
   112|}): Promise<EvidenceSubmission>;
   113|
   114|export function submitJudgment(args: {
   115|  baseUrl: string;
   116|  accessToken: string;
   117|  stuntId: string;
   118|  commitment: number;
   119|  transgression: number;
   120|  creativity: number;
   121|  documentation: number;
   122|  fetchImpl?: typeof fetch;
   123|}): Promise<Judgment>;
   124|