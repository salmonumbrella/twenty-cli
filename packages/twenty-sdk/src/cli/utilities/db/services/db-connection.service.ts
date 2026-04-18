import { Client } from "pg";

export interface DbConnectionOptions {
  databaseUrl: string;
}

export class DbConnectionService {
  async connect(options: DbConnectionOptions): Promise<Client> {
    const url = new URL(options.databaseUrl);
    const sslmode = url.searchParams.get("sslmode");
    const client = new Client({
      connectionString: options.databaseUrl,
      ssl: sslmode === "require" ? { rejectUnauthorized: false } : undefined,
    });

    await client.connect();

    return client;
  }

  async ping(client: Pick<Client, "query">): Promise<{ ok: true }> {
    await client.query("select 1");

    return { ok: true };
  }
}
