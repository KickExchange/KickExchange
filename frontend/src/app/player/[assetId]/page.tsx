"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { graphqlRequest, Asset, PlayerPreview } from "@/lib/graphql";

const PLAYER_QUERY = `
  query Player($assetId: Uint64!) {
    player(assetId: $assetId) {
      assetId
      externalId
      symbol
      displayName
      initialPrice
    }
  }
`;

const PROFILE_QUERY = `
  query PreviewPlayer($query: String!) {
    previewPlayer(query: $query) {
      marketValue
      imageUrl
      position
      club
      nationalities
      shirtNumber
    }
  }
`;

export default function PlayerDetailPage() {
  const params = useParams<{ assetId: string }>();
  const [asset, setAsset] = useState<Asset | null>(null);
  const [profile, setProfile] = useState<PlayerPreview | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const data = await graphqlRequest<{ player: Asset | null }>(PLAYER_QUERY, {
          assetId: Number(params.assetId),
        });
        setAsset(data.player);

        if (data.player) {
          // profile/market value come live from transfermarkt - separate
          // from our platform's trading price (initialPrice / a future
          // last-trade price)
          const preview = await graphqlRequest<{ previewPlayer: PlayerPreview | null }>(
            PROFILE_QUERY,
            { query: data.player.externalId }
          );
          setProfile(preview.previewPlayer);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load player");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [params.assetId]);

  return (
    <div className="flex flex-1 flex-col items-center bg-zinc-50 px-4 py-24 dark:bg-black">
      <div className="flex w-full max-w-xl flex-col gap-6">
        <Link href="/" className="text-sm text-zinc-500 hover:underline dark:text-zinc-400">
          &larr; back to search
        </Link>

        {loading && <p className="text-zinc-600 dark:text-zinc-400">Loading...</p>}
        {error && <p className="text-red-600 dark:text-red-400">{error}</p>}
        {!loading && !error && !asset && (
          <p className="text-zinc-600 dark:text-zinc-400">Player not found.</p>
        )}

        {asset && (
          <div className="overflow-hidden rounded-xl border border-black/10 bg-white dark:border-white/15 dark:bg-zinc-900">
            <div className="flex items-center gap-5 p-8">
              {profile?.imageUrl && (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={profile.imageUrl}
                  alt={asset.displayName}
                  className="h-24 w-24 rounded-full object-cover"
                />
              )}
              <div>
                <h1 className="text-3xl font-semibold tracking-tight text-black dark:text-zinc-50">
                  {asset.displayName}
                </h1>
                <p className="mt-1 text-zinc-500 dark:text-zinc-400">
                  {asset.symbol}
                  {profile?.shirtNumber && ` - ${profile.shirtNumber}`}
                </p>
                {profile && (
                  <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                    {[profile.position, profile.club, profile.nationalities.join(", ")]
                      .filter(Boolean)
                      .join(" - ")}
                  </p>
                )}
              </div>
            </div>

            <div className="flex flex-col gap-3 border-t border-black/10 p-8 dark:border-white/15">
              <div className="flex items-baseline gap-2">
                <span className="text-sm text-zinc-500 dark:text-zinc-400">Market value</span>
                <span className="font-mono text-2xl text-black dark:text-zinc-50">
                  {profile ? `$${profile.marketValue.toLocaleString()}` : "-"}
                </span>
              </div>
              <div className="flex items-baseline gap-2">
                <span className="text-sm text-zinc-500 dark:text-zinc-400">Platform price</span>
                <span className="font-mono text-lg text-black dark:text-zinc-50">
                  ${asset.initialPrice.toFixed(2)}
                </span>
              </div>
              <p className="mt-2 text-xs text-zinc-400 dark:text-zinc-600">
                transfermarkt id {asset.externalId}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
