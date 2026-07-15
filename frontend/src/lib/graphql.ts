const GRAPHQL_URL = process.env.NEXT_PUBLIC_GRAPHQL_URL ?? "http://localhost:8080/graphql";

export class GraphQLRequestError extends Error {}

export async function graphqlRequest<T>(
  query: string,
  variables?: Record<string, unknown>
): Promise<T> {
  const res = await fetch(GRAPHQL_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query, variables }),
  });
  const json = await res.json();
  if (json.errors?.length) {
    throw new GraphQLRequestError(json.errors[0].message);
  }
  return json.data as T;
}

export interface Asset {
  assetId: number;
  externalId: string;
  symbol: string;
  displayName: string;
  initialPrice: number;
}

export interface PlayerPreview {
  externalId: string;
  name: string;
  marketValue: number;
}
