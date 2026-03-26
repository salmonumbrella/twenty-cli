import http from "node:http";
import { AddressInfo } from "node:net";

export interface MockGraphqlRequest {
  pathname: string;
  body: string;
}

export interface MockBinaryRequest {
  method: string;
  path: string;
}

export async function startMockGraphqlServer(
  respond: (body: string) => Record<string, unknown>,
): Promise<{
  baseUrl: string;
  requests: MockGraphqlRequest[];
  close: () => Promise<void>;
}> {
  const requests: MockGraphqlRequest[] = [];
  const server = http.createServer((req, res) => {
    const chunks: Buffer[] = [];

    req.on("data", (chunk) => {
      chunks.push(Buffer.from(chunk));
    });

    req.on("end", () => {
      const body = Buffer.concat(chunks).toString("utf-8");
      requests.push({
        pathname: req.url ?? "",
        body,
      });

      res.statusCode = 200;
      res.setHeader("content-type", "application/json");
      res.setHeader("connection", "close");
      res.end(JSON.stringify(respond(body)));
    });
  });

  await listenOnRandomPort(server);
  const { port } = getBoundPort(server, "mock GraphQL server");

  return {
    baseUrl: `http://127.0.0.1:${port}`,
    requests,
    close: () => closeServer(server),
  };
}

export async function startMockBinaryServer(body: Buffer): Promise<{
  baseUrl: string;
  requests: MockBinaryRequest[];
  close: () => Promise<void>;
}> {
  const requests: MockBinaryRequest[] = [];
  const server = http.createServer((req, res) => {
    requests.push({
      method: req.method ?? "GET",
      path: req.url ?? "",
    });

    res.statusCode = 200;
    res.setHeader("content-type", "application/octet-stream");
    res.setHeader("connection", "close");
    res.end(body);
  });

  await listenOnRandomPort(server);
  const { port } = getBoundPort(server, "mock download server");

  return {
    baseUrl: `http://127.0.0.1:${port}`,
    requests,
    close: () => closeServer(server),
  };
}

export function createLocalRequestEnv(baseUrl?: string): NodeJS.ProcessEnv {
  return {
    ...(baseUrl ? { TWENTY_BASE_URL: baseUrl } : {}),
    TWENTY_NO_RETRY: "true",
    NODE_OPTIONS: "",
    HTTP_PROXY: "",
    HTTPS_PROXY: "",
    ALL_PROXY: "",
    NO_PROXY: "127.0.0.1,localhost",
    VITEST: "",
    VITEST_WORKER_ID: "",
    VITEST_POOL_ID: "",
  };
}

async function listenOnRandomPort(server: http.Server): Promise<void> {
  await new Promise<void>((resolve) => {
    server.listen(0, "127.0.0.1", resolve);
  });
  server.unref();
}

function getBoundPort(server: http.Server, name: string): AddressInfo {
  const address = server.address();

  if (!address || typeof address === "string") {
    throw new Error(`Failed to bind ${name}`);
  }

  return address as AddressInfo;
}

function closeServer(server: http.Server): Promise<void> {
  server.close();
  return Promise.resolve();
}
