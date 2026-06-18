export async function getMe({ baseUrl, accessToken, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/me`, { headers: authHeaders(accessToken) });

  if (!response.ok) {
    throw new Error(`getMe failed with status ${response.status}`);
  }

  return response.json();
}

export async function updateDisplayName({ baseUrl, accessToken, displayName, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/me/display-name`, {
    method: "PATCH",
    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    body: JSON.stringify({ displayName }),
  });

  if (!response.ok) {
    throw new Error(`updateDisplayName failed with status ${response.status}`);
  }

  return response.json();
}

export async function createJump({
  baseUrl,
  accessToken,
  source,
  destination,
  food,
  caption,
  mediaObjectKey,
  fetchImpl = fetch,
}) {
  const response = await fetchImpl(`${baseUrl}/v1/jumps`, {
    method: "POST",
    headers: { ...authHeaders(accessToken), "Content-Type": "application/json" },
    body: JSON.stringify({ source, destination, food, caption, mediaObjectKey }),
  });

  if (!response.ok) {
    throw new Error(`createJump failed with status ${response.status}`);
  }

  return response.json();
}

export async function submitJudgment({
  baseUrl,
  accessToken,
  guestSessionId,
  jumpId,
  commitment,
  transgression,
  creativity,
  presentation,
  fetchImpl = fetch,
}) {
  const headers = { "Content-Type": "application/json", Accept: "application/json" };
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }

  const response = await fetchImpl(`${baseUrl}/v1/jumps/${jumpId}/judgment`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      guestSessionId,
      commitment,
      transgression,
      creativity,
      presentation,
    }),
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, `submitJudgment failed with status ${response.status}`));
  }

  return response.json();
}

export async function createGuestSession({ baseUrl, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/guest-sessions`, {
    method: "POST",
    headers: { Accept: "application/json" },
  });

  if (!response.ok) {
    throw new Error(await errorMessage(response, `createGuestSession failed with status ${response.status}`));
  }

  return response.json();
}

export async function getPublicFeed({ baseUrl, accessToken, cursor, limit = 20, fetchImpl = fetch }) {
  const params = new URLSearchParams();
  if (cursor) params.set("cursor", cursor);
  params.set("limit", String(limit));

  const headers = { Accept: "application/json" };
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }

  const response = await fetchImpl(`${baseUrl}/v1/feed?${params.toString()}`, { headers });

  if (!response.ok) {
    throw new Error(await errorMessage(response, `getPublicFeed failed with status ${response.status}`));
  }

  return response.json();
}

export async function getJumpDetail({ baseUrl, accessToken, jumpId, fetchImpl = fetch }) {
  const headers = { Accept: "application/json" };
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }

  const response = await fetchImpl(`${baseUrl}/v1/jumps/${jumpId}`, { headers });

  if (!response.ok) {
    throw new Error(await errorMessage(response, `getJumpDetail failed with status ${response.status}`));
  }

  return response.json();
}

function authHeaders(accessToken) {
  return {
    Authorization: `Bearer ${accessToken}`,
    Accept: "application/json",
  };
}

async function errorMessage(response, fallback) {
  const body = await response.text();
  if (body) {
    try {
      const parsed = JSON.parse(body);
      if (parsed && typeof parsed.message === "string" && parsed.message) {
        return parsed.message;
      }
    } catch {
      return body;
    }
  }
  return fallback;
}
