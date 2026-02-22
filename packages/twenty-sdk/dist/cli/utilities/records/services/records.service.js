"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.RecordsService = void 0;
const parse_1 = require("../../shared/parse");
class RecordsService {
    constructor(api) {
        this.api = api;
    }
    async list(object, options = {}) {
        const params = {};
        if (options.limit)
            params.limit = String(options.limit);
        if (options.cursor)
            params.starting_after = options.cursor;
        if (options.sort)
            params.order_by = options.sort;
        if (options.order)
            params.order_by_direction = options.order;
        if (options.fields)
            params.fields = options.fields;
        if (options.include)
            params.depth = '1';
        if (options.filter)
            params.filter = options.filter;
        if (options.params) {
            for (const [key, values] of Object.entries(options.params)) {
                params[key] = values.length === 1 ? values[0] : values;
            }
        }
        const response = await this.api.get(`/rest/${object}`, { params });
        const payload = response.data;
        const dataSection = payload?.data ?? {};
        const records = extractArray(dataSection, object);
        return {
            data: records,
            totalCount: payload?.totalCount,
            pageInfo: payload?.pageInfo,
        };
    }
    async listAll(object, options = {}) {
        const all = [];
        let cursor = options.cursor ?? '';
        let pageInfo;
        let totalCount;
        while (true) {
            const response = await this.list(object, { ...options, cursor });
            all.push(...response.data);
            pageInfo = response.pageInfo;
            totalCount = response.totalCount ?? totalCount;
            if (!pageInfo?.hasNextPage || !pageInfo?.endCursor) {
                break;
            }
            cursor = pageInfo.endCursor;
        }
        return { data: all, totalCount, pageInfo };
    }
    async get(object, id, options) {
        const params = {};
        if (options?.include) {
            params.depth = '1';
        }
        const response = await this.api.get(`/rest/${object}/${id}`, { params });
        const payload = response.data;
        const dataSection = payload?.data ?? {};
        const singular = (0, parse_1.singularize)(object);
        return dataSection[singular] ?? dataSection[object] ?? firstValue(dataSection);
    }
    async create(object, data) {
        const response = await this.api.post(`/rest/${object}`, data);
        const payload = response.data;
        const dataSection = payload?.data ?? {};
        const key = `create${(0, parse_1.capitalize)((0, parse_1.singularize)(object))}`;
        return dataSection[key] ?? firstValue(dataSection);
    }
    async update(object, id, data) {
        const response = await this.api.patch(`/rest/${object}/${id}`, data);
        const payload = response.data;
        const dataSection = payload?.data ?? {};
        const key = `update${(0, parse_1.capitalize)((0, parse_1.singularize)(object))}`;
        return dataSection[key] ?? firstValue(dataSection);
    }
    async delete(object, id) {
        const response = await this.api.delete(`/rest/${object}/${id}`);
        return response.data ?? null;
    }
    async destroy(object, id) {
        const response = await this.api.delete(`/rest/${object}/${id}/destroy`);
        return response.data ?? null;
    }
    async restore(object, id) {
        const response = await this.api.post(`/rest/${object}/${id}/restore`);
        return response.data ?? null;
    }
    async batchCreate(object, records) {
        const response = await this.api.post(`/rest/batch/${object}`, records);
        return response.data ?? null;
    }
    async batchUpdate(object, records) {
        const response = await this.api.patch(`/rest/batch/${object}`, records);
        return response.data ?? null;
    }
    async batchDelete(object, ids) {
        const filter = `id[in]:[${ids.join(',')}]`;
        const response = await this.api.delete(`/rest/batch/${object}`, { params: { filter } });
        return response.data ?? null;
    }
    async groupBy(object, payload, params) {
        const path = `/rest/${object}/group-by`;
        if (payload) {
            const response = await this.api.post(path, payload);
            return response.data ?? null;
        }
        const response = await this.api.get(path, { params: flattenParams(params) });
        return response.data ?? null;
    }
    async findDuplicates(object, payload) {
        const response = await this.api.post(`/rest/${object}/find-duplicates`, payload);
        return response.data ?? null;
    }
    async merge(object, payload) {
        const response = await this.api.patch(`/rest/${object}/merge`, payload);
        return response.data ?? null;
    }
}
exports.RecordsService = RecordsService;
function extractArray(dataSection, object) {
    const raw = dataSection?.[object];
    if (Array.isArray(raw))
        return raw;
    for (const value of Object.values(dataSection)) {
        if (Array.isArray(value))
            return value;
    }
    return [];
}
function firstValue(dataSection) {
    const values = Object.values(dataSection);
    if (values.length === 0)
        return undefined;
    return values[0];
}
function flattenParams(params) {
    if (!params)
        return undefined;
    const out = {};
    for (const [key, values] of Object.entries(params)) {
        out[key] = values.length === 1 ? values[0] : values;
    }
    return out;
}
