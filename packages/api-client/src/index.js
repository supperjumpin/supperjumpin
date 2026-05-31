export async function getMe({ baseUrl, accessToken, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/me`, { headers: authHeaders(accessToken) });

  if (!response.ok) {
    throw new Error(`getMe failed with status ${response.status}`);
  }

  return response.json();
}

export async function createGroup({ baseUrl, accessToken, name, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/groups`, {
    method: "POST",
    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });

  if (!response.ok) {
    throw new Error(`createGroup failed with status ${response.status}`);
  }

  return response.json();
}

export async function listGroups({ baseUrl, accessToken, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/groups`, { headers: authHeaders(accessToken) });

  if (!response.ok) {
    throw new Error(`listGroups failed with status ${response.status}`);
  }

  return response.json();
}

export async function getGroupHome({ baseUrl, accessToken, groupId, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/home`, {
    headers: authHeaders(accessToken),
  });

  if (!response.ok) {
    throw new Error(`getGroupHome failed with status ${response.status}`);
  }

  return response.json();
}

export async function createInvite({ baseUrl, accessToken, groupId, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/invites`, {
    method: "POST",
    headers: authHeaders(accessToken),
  });

  if (!response.ok) {
    throw new Error(`createInvite failed with status ${response.status}`);
  }

  return response.json();
}

export async function acceptInvite({ baseUrl, accessToken, token, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/invites/${token}/accept`, {
    method: "POST",
    headers: authHeaders(accessToken),
  });

  if (!response.ok) {
    throw new Error(`acceptInvite failed with status ${response.status}`);
  }

  return response.json();
}

export async function startSeason({ baseUrl, accessToken, groupId, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/seasons`, {
    method: "POST",
    headers: authHeaders(accessToken),
  });

  if (!response.ok) {
    throw new Error(`startSeason failed with status ${response.status}`);
  }

  return response.json();
}

export async function createIdea({ baseUrl, accessToken, groupId, source, destination, food, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/groups/${groupId}/ideas`, {
    method: "POST",
    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    body: JSON.stringify({ source, destination, food }),
  });

  if (!response.ok) {
    throw new Error(`createIdea failed with status ${response.status}`);
  }

  return response.json();
}

export async function createPlannedJump({ baseUrl, accessToken, ideaId, offSeason = false, fetchImpl = fetch }) {
  const init = {
    method: "POST",
    headers: authHeaders(accessToken),
  };
  if (offSeason) {
    init.headers = { ...init.headers, "Content-Type": "application/json" };
    init.body = JSON.stringify({ offSeason: true });
  }

  const response = await fetchImpl(`${baseUrl}/v1/ideas/${ideaId}/planned-jump`, init);

  if (!response.ok) {
    throw new Error(`createPlannedJump failed with status ${response.status}`);
  }

  return response.json();
}

export async function authorizeEvidenceUpload({ baseUrl, accessToken, jumpId, contentType, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/jumps/${jumpId}/evidence-upload-authorizations`, {
    method: "POST",
    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    body: JSON.stringify({ contentType }),
  });

  if (!response.ok) {
    throw new Error(`authorizeEvidenceUpload failed with status ${response.status}`);
  }

  return response.json();
}

export async function submitEvidence({ baseUrl, accessToken, jumpId, uploadAuthorizationId, caption, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/jumps/${jumpId}/evidence`, {
    method: "POST",
    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    body: JSON.stringify({ uploadAuthorizationId, caption }),
  });

  if (!response.ok) {
    throw new Error(`submitEvidence failed with status ${response.status}`);
  }

  return response.json();
}

export async function submitJudgment({
  baseUrl,
  accessToken,
  jumpId,
  difficulty,
  transgression,
  creativity,
  presentation,
  fetchImpl = fetch,
}) {
  const response = await fetchImpl(`${baseUrl}/v1/jumps/${jumpId}/judgment`, {
    method: "POST",
    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    body: JSON.stringify({ difficulty, transgression, creativity, presentation }),
  });

  if (!response.ok) {
    throw new Error(`submitJudgment failed with status ${response.status}`);
  }

  return response.json();
}

function authHeaders(accessToken) {
  return {
    Authorization: `Bearer ${accessToken}`,
    Accept: "application/json",
  };
}
