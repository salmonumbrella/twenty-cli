import { ApiService } from '../../api/services/api.service';

export interface ObjectMetadata {
  id: string;
  nameSingular?: string;
  namePlural?: string;
  fields?: FieldMetadata[];
  [key: string]: unknown;
}

export interface FieldMetadata {
  id: string;
  objectMetadataId?: string;
  [key: string]: unknown;
}

export class MetadataService {
  constructor(private api: ApiService) {}

  async listObjects(): Promise<ObjectMetadata[]> {
    const response = await this.api.get('/rest/metadata/objects');
    const payload = response.data as any;
    return payload?.data?.objects ?? [];
  }

  async getObject(nameOrId: string): Promise<ObjectMetadata> {
    if (looksLikeUuid(nameOrId)) {
      const response = await this.api.get(`/rest/metadata/objects/${nameOrId}`);
      const payload = response.data as any;
      return payload?.data?.object ?? payload?.data ?? {};
    }

    const objects = await this.listObjects();
    const match = objects.find((obj) =>
      obj.nameSingular === nameOrId || obj.namePlural === nameOrId
    );
    if (!match) {
      throw new Error(`Object not found: ${nameOrId}`);
    }

    const response = await this.api.get(`/rest/metadata/objects/${match.id}`);
    const payload = response.data as any;
    return payload?.data?.object ?? payload?.data ?? {};
  }

  async createObject(data: Record<string, unknown>): Promise<unknown> {
    const response = await this.api.post('/rest/metadata/objects', data);
    return response.data ?? null;
  }

  async listFields(): Promise<FieldMetadata[]> {
    const response = await this.api.get('/rest/metadata/fields');
    const payload = response.data as any;
    return payload?.data?.fields ?? [];
  }

  async getField(id: string): Promise<FieldMetadata> {
    const response = await this.api.get(`/rest/metadata/fields/${id}`);
    const payload = response.data as any;
    return payload?.data?.field ?? payload?.data ?? {};
  }

  async createField(data: Record<string, unknown>): Promise<unknown> {
    const response = await this.api.post('/rest/metadata/fields', data);
    return response.data ?? null;
  }
}

function looksLikeUuid(value: string): boolean {
  return value.length === 36 && value[8] === '-' && value[13] === '-';
}
