"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.MetadataService = void 0;
class MetadataService {
    constructor(api) {
        this.api = api;
    }
    async listObjects() {
        const response = await this.api.get('/rest/metadata/objects');
        const payload = response.data;
        return payload?.data?.objects ?? [];
    }
    async getObject(nameOrId) {
        if (looksLikeUuid(nameOrId)) {
            const response = await this.api.get(`/rest/metadata/objects/${nameOrId}`);
            const payload = response.data;
            return payload?.data?.object ?? payload?.data ?? {};
        }
        const objects = await this.listObjects();
        const match = objects.find((obj) => obj.nameSingular === nameOrId || obj.namePlural === nameOrId);
        if (!match) {
            throw new Error(`Object not found: ${nameOrId}`);
        }
        const response = await this.api.get(`/rest/metadata/objects/${match.id}`);
        const payload = response.data;
        return payload?.data?.object ?? payload?.data ?? {};
    }
    async createObject(data) {
        const response = await this.api.post('/rest/metadata/objects', data);
        return response.data ?? null;
    }
    async listFields() {
        const response = await this.api.get('/rest/metadata/fields');
        const payload = response.data;
        return payload?.data?.fields ?? [];
    }
    async getField(id) {
        const response = await this.api.get(`/rest/metadata/fields/${id}`);
        const payload = response.data;
        return payload?.data?.field ?? payload?.data ?? {};
    }
    async createField(data) {
        const response = await this.api.post('/rest/metadata/fields', data);
        return response.data ?? null;
    }
}
exports.MetadataService = MetadataService;
function looksLikeUuid(value) {
    return value.length === 36 && value[8] === '-' && value[13] === '-';
}
