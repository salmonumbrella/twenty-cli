const RELEVANT_CHANGE_PATTERNS = [
  /^packages\/twenty-server\/src\/engine\/api\//,
  /^packages\/twenty-server\/src\/engine\/metadata-modules\//,
  /^packages\/twenty-server\/src\/engine\/workspace-manager\//,
  /^packages\/twenty-server\/src\/modules\/.+\/controllers?\//,
  /^packages\/twenty-server\/src\/modules\/.+\/resolvers?\//,
  /^packages\/twenty-server\/src\/modules\/.+\/dtos?\//,
  /^packages\/twenty-server\/src\/database\/commands\/upgrade-version-command\//,
  /^packages\/twenty-server\/src\/database\/typeorm\/core\/migrations\/common\//,
];

function normalizeAuditSha(auditSha) {
  if (!/^[0-9a-f]{40}$/.test(auditSha ?? "")) {
    throw new Error("Expected auditSha to be a 40-character lowercase hex commit SHA");
  }

  return auditSha;
}

export function classifyRelevantUpstreamChanges(files) {
  return files.filter((file) => RELEVANT_CHANGE_PATTERNS.some((pattern) => pattern.test(file)));
}

function buildHeaders(token) {
  const headers = {
    Accept: "application/vnd.github+json",
    "User-Agent": "twenty-cli-upstream-drift-check",
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  return headers;
}

async function githubJson(url, token, fetchImpl = fetch) {
  const response = await fetchImpl(url, {
    headers: buildHeaders(token),
  });

  if (!response.ok) {
    throw new Error(`GitHub API request failed: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export async function fetchLatestSha({ token, fetchImpl } = {}) {
  const body = await githubJson(
    "https://api.github.com/repos/twentyhq/twenty/commits/main",
    token,
    fetchImpl,
  );

  return body.sha;
}

export async function fetchChangedFilesSinceAudit({ auditSha, latestSha, token, fetchImpl } = {}) {
  const body = await githubJson(
    `https://api.github.com/repos/twentyhq/twenty/compare/${auditSha}...${latestSha}`,
    token,
    fetchImpl,
  );

  return (body.files ?? []).map((file) => file.filename);
}

export async function checkUpstreamDrift({ auditSha, token, fetchImpl } = {}) {
  const normalizedAuditSha = normalizeAuditSha(auditSha);
  const latestSha = await fetchLatestSha({ token, fetchImpl });

  if (latestSha === normalizedAuditSha) {
    return {
      auditSha: normalizedAuditSha,
      latestSha,
      relevantFiles: [],
      status: "current",
    };
  }

  const changedFiles = await fetchChangedFilesSinceAudit({
    auditSha: normalizedAuditSha,
    latestSha,
    token,
    fetchImpl,
  });
  const relevantFiles = classifyRelevantUpstreamChanges(changedFiles);

  if (relevantFiles.length > 0) {
    return {
      auditSha: normalizedAuditSha,
      changedFiles,
      latestSha,
      relevantFiles,
      status: "relevant_drift",
    };
  }

  return {
    auditSha: normalizedAuditSha,
    changedFiles,
    latestSha,
    relevantFiles: [],
    status: "non_api_drift",
  };
}
