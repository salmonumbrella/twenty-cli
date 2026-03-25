import { ApiService } from "../../api/services/api.service";

type SearchApiClient = Pick<ApiService, "post">;

export interface SearchResult {
  recordId: string;
  objectNameSingular: string;
  objectLabelSingular: string;
  label: string;
  imageUrl?: string | null;
  tsRankCD: number;
  tsRank: number;
  cursor?: string;
}

export interface SearchPageInfo {
  hasNextPage?: boolean;
  endCursor?: string;
}

export interface SearchResponse {
  data: SearchResult[];
  pageInfo?: SearchPageInfo;
}

interface GraphQLSearchResponse {
  data?: {
    search?: {
      edges?: Array<{
        cursor: string;
        node: SearchResult;
      }>;
      pageInfo?: SearchPageInfo;
    };
  };
}

export class SearchService {
  constructor(private api: SearchApiClient) {}

  async search(options: {
    query: string;
    limit?: number;
    objects?: string[];
    excludeObjects?: string[];
    after?: string;
    filter?: Record<string, unknown>;
  }): Promise<SearchResponse> {
    const query = `
      query Search($searchInput: String!, $limit: Int, $after: String, $filter: ObjectRecordFilterInput, $includedObjectNameSingulars: [String!], $excludedObjectNameSingulars: [String!]) {
        search(
          searchInput: $searchInput
          limit: $limit
          after: $after
          filter: $filter
          includedObjectNameSingulars: $includedObjectNameSingulars
          excludedObjectNameSingulars: $excludedObjectNameSingulars
        ) {
          edges {
            cursor
            node {
              recordId
              objectNameSingular
              objectLabelSingular
              label
              imageUrl
              tsRankCD
              tsRank
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
      }
    `;
    const variables = {
      searchInput: options.query,
      limit: options.limit,
      after: options.after,
      filter: options.filter,
      includedObjectNameSingulars: options.objects,
      excludedObjectNameSingulars: options.excludeObjects,
    };
    const response = await this.api.post<GraphQLSearchResponse>("/graphql", { query, variables });
    return {
      data:
        response.data?.data?.search?.edges?.map((edge) => ({
          ...edge.node,
          cursor: edge.cursor,
        })) ?? [],
      pageInfo: response.data?.data?.search?.pageInfo,
    };
  }
}
