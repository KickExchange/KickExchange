"use client";

import { useState } from "react";
import { graphqlRequest, Asset } from "@/lib/graphql";

const SEARCH_QUERY = `
  query SearchAssets($query: String!) {
    searchAssets(query: $query) {
      assetId
      symbol
      displayName
      initialPrice
    }
  }
`;

export default function Home() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Asset[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searched, setSearched] = useState(false);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim()) return;
    setLoading(true);
    setError(null);
    try {
      const data = await graphqlRequest<{ searchAssets: Asset[] }>(SEARCH_QUERY, { query });
      setResults(data.searchAssets);
      setSearched(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Search failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-col flex-1 items-center bg-zinc-50 px-4 py-24 dark:bg-black">
      <div className="flex w-full max-w-xl flex-col items-center gap-8">
        <h1 className="text-4xl font-semibold tracking-tight text-black dark:text-zinc-50">
          KickExchange
        </h1>

        <form onSubmit={handleSearch} className="flex w-full gap-2">
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search a player..."
            className="flex-1 rounded-full border border-black/10 bg-white px-5 py-3 text-black outline-none focus:border-black/30 dark:border-white/15 dark:bg-zinc-900 dark:text-zinc-50 dark:focus:border-white/30"
          />
          <button
            type="submit"
            disabled={loading}
            className="rounded-full bg-foreground px-6 py-3 font-medium text-background hover:bg-[#383838] disabled:opacity-50 dark:hover:bg-[#ccc]"
          >
            {loading ? "..." : "Search"}
          </button>
        </form>

        {error && <p className="text-red-600 dark:text-red-400">{error}</p>}

        {searched && !error && results.length === 0 && (
          <p className="text-zinc-600 dark:text-zinc-400">
            No tradable players found for &quot;{query}&quot;.
          </p>
        )}

        {results.length > 0 && (
          <ul className="flex w-full flex-col gap-3">
            {results.map((a) => (
              <li
                key={a.assetId}
                className="flex items-center justify-between rounded-xl border border-black/10 bg-white px-5 py-4 dark:border-white/15 dark:bg-zinc-900"
              >
                <div>
                  <p className="font-medium text-black dark:text-zinc-50">{a.displayName}</p>
                  <p className="text-sm text-zinc-500 dark:text-zinc-400">{a.symbol}</p>
                </div>
                <p className="font-mono text-black dark:text-zinc-50">${a.initialPrice.toFixed(2)}</p>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
