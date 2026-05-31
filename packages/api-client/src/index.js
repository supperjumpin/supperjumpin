     1|export async function getMe({ baseUrl, accessToken, fetchImpl = fetch }) {
     2|  const response = await fetchImpl(`${baseUrl}/v1/me`, { headers: authHeaders(accessToken) });
     3|
     4|  if (!response.ok) {
     5|    throw new Error(`getMe failed with status ${response.status}`);
     6|  }
     7|
     8|  return response.json();
     9|}
    10|
    11|export async function createGroup({ baseUrl, accessToken, name, fetchImpl = fetch }) {
    12|  const response = await fetchImpl(`${baseUrl}/v1/groups`, {
    13|    method: "POST",
    14|    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    15|    body: JSON.stringify({ name }),
    16|  });
    17|
    18|  if (!response.ok) {
    19|    throw new Error(`createGroup failed with status ${response.status}`);
    20|  }
    21|
    22|  return response.json();
    23|}
    24|
    25|export async function listGroups({ baseUrl, accessToken, fetchImpl = fetch }) {
    26|  const response = await fetchImpl(`${baseUrl}/v1/groups`, { headers: authHeaders(accessToken) });
    27|
    28|  if (!response.ok) {
    29|    throw new Error(`listGroups failed with status ${response.status}`);
    30|  }
    31|
    32|  return response.json();
    33|}
    34|
    35|export async function getGroupHome({ baseUrl, accessToken, groupId, fetchImpl = fetch }) {
    36|  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/home`, {
    37|    headers: authHeaders(accessToken),
    38|  });
    39|
    40|  if (!response.ok) {
    41|    throw new Error(`getGroupHome failed with status ${response.status}`);
    42|  }
    43|
    44|  return response.json();
    45|}
    46|
    47|export async function createInvite({ baseUrl, accessToken, groupId, fetchImpl = fetch }) {
    48|  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/invites`, {
    49|    method: "POST",
    50|    headers: authHeaders(accessToken),
    51|  });
    52|
    53|  if (!response.ok) {
    54|    throw new Error(`createInvite failed with status ${response.status}`);
    55|  }
    56|
    57|  return response.json();
    58|}
    59|
    60|export async function acceptInvite({ baseUrl, accessToken, token, fetchImpl = fetch }) {
    61|  const response = await fetchImpl(`${baseUrl}/v1/invites/${token}/accept`, {
    62|    method: "POST",
    63|    headers: authHeaders(accessToken),
    64|  });
    65|
    66|  if (!response.ok) {
    67|    throw new Error(`acceptInvite failed with status ${response.status}`);
    68|  }
    69|
    70|  return response.json();
    71|}
    72|
    73|export async function startSeason({ baseUrl, accessToken, groupId, fetchImpl = fetch }) {
    74|  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/seasons`, {
    75|    method: "POST",
    76|    headers: authHeaders(accessToken),
    77|  });
    78|
    79|  if (!response.ok) {
    80|    throw new Error(`startSeason failed with status ${response.status}`);
    81|  }
    82|
    83|  return response.json();
    84|}
    85|
    86|export async function createIdea({ baseUrl, accessToken, groupId, source, destination, food, fetchImpl = fetch }) {
    87|  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/ideas`, {
    88|    method: "POST",
    89|    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    90|    body: JSON.stringify({ source, destination, food }),
    91|  });
    92|
    93|  if (!response.ok) {
    94|    throw new Error(`createIdea failed with status ${response.status}`);
    95|  }
    96|
    97|  return response.json();
    98|}
    99|
   100|export async function createPlannedStunt({ baseUrl, accessToken, ideaId, offSeason = false, fetchImpl = fetch }) {
   101|  const init = {
   102|    method: "POST",
   103|    headers: authHeaders(accessToken),
   104|  };
   105|  if (offSeason) {
   106|    init.headers = { ...init.headers, "Content-Type": "application/json" };
   107|    init.body = JSON.stringify({ offSeason: true });
   108|  }
   109|
   110|  const response = await fetchImpl(`${baseUrl}/v1/ideas/${ideaId}/planned-stunt`, init);
   111|
   112|  if (!response.ok) {
   113|    throw new Error(`createPlannedStunt failed with status ${response.status}`);
   114|  }
   115|
   116|  return response.json();
   117|}
   118|
   119|export async function authorizeEvidenceUpload({ baseUrl, accessToken, stuntId, contentType, fetchImpl = fetch }) {
   120|  const response = await fetchImpl(`${baseUrl}/v1/stunts/${stuntId}/evidence-upload-authorizations`, {
   121|    method: "POST",
   122|    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
   123|    body: JSON.stringify({ contentType }),
   124|  });
   125|
   126|  if (!response.ok) {
   127|    throw new Error(`authorizeEvidenceUpload failed with status ${response.status}`);
   128|  }
   129|
   130|  return response.json();
   131|}
   132|
   133|export async function submitEvidence({ baseUrl, accessToken, stuntId, uploadAuthorizationId, caption, fetchImpl = fetch }) {
   134|  const response = await fetchImpl(`${baseUrl}/v1/stunts/${stuntId}/evidence`, {
   135|    method: "POST",
   136|    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
   137|    body: JSON.stringify({ uploadAuthorizationId, caption }),
   138|  });
   139|
   140|  if (!response.ok) {
   141|    throw new Error(`submitEvidence failed with status ${response.status}`);
   142|  }
   143|
   144|  return response.json();
   145|}
   146|
   147|export async function submitJudgment({
   148|  baseUrl,
   149|  accessToken,
   150|  stuntId,
   151|  commitment,
   152|  transgression,
   153|  creativity,
   154|  documentation,
   155|  fetchImpl = fetch,
   156|}) {
   157|  const response = await fetchImpl(`${baseUrl}/v1/stunts/${stuntId}/judgment`, {
   158|    method: "POST",
   159|    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
   160|    body: JSON.stringify({ commitment, transgression, creativity, documentation }),
   161|  });
   162|
   163|  if (!response.ok) {
   164|    throw new Error(`submitJudgment failed with status ${response.status}`);
   165|  }
   166|
   167|  return response.json();
   168|}
   169|
   170|function authHeaders(accessToken) {
   171|  return {
   172|    Authorization: `Bearer ${accessToken}`,
   173|    Accept: "application/json",
   174|  };
   175|}
   176|