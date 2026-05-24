export async function getMe({ baseUrl, accessToken, fetchImpl = fetch }) {
  const response = await fetchImpl(`${baseUrl}/v1/me`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`getMe failed with status ${response.status}`);
  }

  return response.json();
}
