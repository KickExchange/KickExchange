"use client";

import { useState } from "react";
import { graphqlRequest, Asset, PlayerPreview } from "@/lib/graphql";

const PREVIEW_QUERY = `
  query PreviewPlayer($query: String!) {
    previewPlayer(query: $query) {
      externalId
      name
      marketValue
    }
  }
`;

const ADD_PLAYER_MUTATION = `
  mutation AddPlayer($externalId: String!) {
    addPlayer(externalId: $externalId) {
      assetId
      symbol
      displayName
      initialPrice
    }
  }
`;

export default function AddPlayerPage() {
  const [query, setQuery] = useState("");
  const [preview, setPreview] = useState<PlayerPreview | null>(null);
  const [added, setAdded] = useState<Asset | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  async function handlePreview(e: React.FormEvent) {
    e.preventDefault();
    if (!query.trim()) return;
    setLoading(true);
    setError(null);
    setAdded(null);
    setPreview(null);
    try {
      const data = await graphqlRequest<{ previewPlayer: PlayerPreview | null }>(
        PREVIEW_QUERY,
        { query }
      );
      setPreview(data.previewPlayer);
      setSearched(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lookup failed");
    } finally {
      setLoading(false);
    }
  }

  async function handleConfirm() {
    if (!preview) return;
    setLoading(true);
    setError(null);
    try {
      const data = await graphqlRequest<{ addPlayer: Asset }>(ADD_PLAYER_MUTATION, {
        externalId: preview.externalId,
      });
      setAdded(data.addPlayer);
      setPreview(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add player");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex flex-1 flex-col items-center bg-zinc-50 px-4 py-24 dark:bg-black">
      <div className="flex w-full max-w-xl flex-col items-center gap-8">
        <h1 className="text-3xl font-semibold tracking-tight text-black dark:text-zinc-50">
          Add player
        </h1>

        <form onSubmit={handlePreview} className="flex w-full gap-2">
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Player name or transfermarkt id..."
            className="flex-1 rounded-full border border-black/10 bg-white px-5 py-3 text-black outline-none focus:border-black/30 dark:border-white/15 dark:bg-zinc-900 dark:text-zinc-50 dark:focus:border-white/30"
          />
          <button
            type="submit"
            disabled={loading}
            className="rounded-full bg-foreground px-6 py-3 font-medium text-background hover:bg-[#383838] disabled:opacity-50 dark:hover:bg-[#ccc]"
          >
            {loading ? "..." : "Look up"}
          </button>
        </form>

        {error && <p className="text-red-600 dark:text-red-400">{error}</p>}

        {searched && !preview && !added && !error && (
          <p className="text-zinc-600 dark:text-zinc-400">
            No player found for &quot;{query}&quot;.
          </p>
        )}

        {preview && (
          <div className="w-full rounded-xl border border-black/10 bg-white p-6 dark:border-white/15 dark:bg-zinc-900">
            <p className="font-medium text-black dark:text-zinc-50">{preview.name}</p>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              transfermarkt id {preview.externalId} - market value $
              {(preview.marketValue / 1_000_000).toFixed(2)}m
            </p>
            <button
              onClick={handleConfirm}
              disabled={loading}
              className="mt-4 rounded-full bg-foreground px-6 py-2 font-medium text-background hover:bg-[#383838] disabled:opacity-50 dark:hover:bg-[#ccc]"
            >
              {loading ? "Adding..." : "Confirm - add to market"}
            </button>
          </div>
        )}

        {added && (
          <div className="w-full rounded-xl border border-green-600/20 bg-green-50 p-6 dark:border-green-400/20 dark:bg-green-950">
            <p className="font-medium text-black dark:text-zinc-50">
              {added.displayName} added - asset #{added.assetId}
            </p>
            <p className="text-sm text-zinc-600 dark:text-zinc-400">
              {added.symbol} - initial price ${added.initialPrice.toFixed(2)}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
