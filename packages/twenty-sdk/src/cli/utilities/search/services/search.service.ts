import { ApiService } from '../../api/services/api.service';

export interface SearchResult {
  recordId: string;
  objectNameSingular: string;
  record: Record<string, unknown>;
}

interface GraphQLSearchResponse {
  data?: {
    search?: SearchResult[];
  };
}

export class SearchService {
  constructor(private api: ApiService) {}

  async search(options: {
    query: string;
    limit?: number;
    objects?: string[];
    excludeObjects?: string[];
  }): Promise<SearchResult[]> {
    const query = `
      query Search($searchInput: String!, $limit: Int, $includedObjectNameSingulars: [String!], $excludedObjectNameSingulars: [String!]) {
        search(
          searchInput: $searchInput
          limit: $limit
          includedObjectNameSingulars: $includedObjectNameSingulars
          excludedObjectNameSingulars: $excludedObjectNameSingulars
        ) {
          recordId
          objectNameSingular
          record
        }
      }
    `;
    const variables = {
      searchInput: options.query,
      limit: options.limit,
      includedObjectNameSingulars: options.objects,
      excludedObjectNameSingulars: options.excludeObjects,
    };
    const response = await this.api.post<GraphQLSearchResponse>('/graphql', { query, variables });
    return response.data?.data?.search ?? [];
  }
}
