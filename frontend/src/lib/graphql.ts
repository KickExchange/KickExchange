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
  imageUrl: string;
  position: string;
  club: string;
  nationalities: string[];
  shirtNumber: string;
}

export interface Transfer {
  date: string;
  season: string;
  clubFromName: string;
  clubToName: string;
  marketValue: number;
}

export interface PlayerProfile {
  externalId: string;
  name: string;
  description: string;
  nameInHomeCountry: string;
  imageUrl: string;
  placeOfBirthCity: string;
  placeOfBirthCountry: string;
  height: number;
  citizenship: string[];
  position: string;
  positionOther: string[];
  foot: string;
  shirtNumber: string;
  clubName: string;
  clubJoined: string;
  clubContractExpires: string;
  marketValue: number;
  agentName: string;
  outfitter: string;
  transfers: Transfer[];
}
