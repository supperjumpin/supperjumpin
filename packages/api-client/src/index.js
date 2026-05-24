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

function authHeaders(accessToken) {
  return {
    Authorization: `Bearer ${accessToken}`,
    Accept: "application/json",
  };
}
