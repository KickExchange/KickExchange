"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { graphqlRequest, Asset, PlayerProfile } from "@/lib/graphql";

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
  query PlayerProfile($externalId: String!) {
    playerProfile(externalId: $externalId) {
      name
      description
      nameInHomeCountry
      imageUrl
      placeOfBirthCity
      placeOfBirthCountry
      height
      citizenship
      position
      positionOther
      foot
      shirtNumber
      clubName
      clubJoined
      clubContractExpires
      marketValue
      agentName
      outfitter
      transfers {
        date
        season
        clubFromName
        clubToName
        marketValue
      }
    }
  }
`;

function money(n: number) {
  return `$${n.toLocaleString()}`;
}

function StatBox({ label, value }: { label: string; value: React.ReactNode }) {
  if (!value) return null;
  return (
    <div className="rounded-lg border border-black/10 bg-white p-4 dark:border-white/15 dark:bg-zinc-900">
      <p className="text-xs uppercase tracking-wide text-zinc-500 dark:text-zinc-400">{label}</p>
      <p className="mt-1 font-medium text-black dark:text-zinc-50">{value}</p>
    </div>
  );
}

export default function PlayerDetailPage() {
  const params = useParams<{ assetId: string }>();
  const [asset, setAsset] = useState<Asset | null>(null);
  const [profile, setProfile] = useState<PlayerProfile | null>(null);
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
          // everything here is live from transfermarkt - separate from our
          // platform's own trading price (initialPrice / a future last-trade price)
          const res = await graphqlRequest<{ playerProfile: PlayerProfile | null }>(
            PROFILE_QUERY,
            { externalId: data.player.externalId }
          );
          setProfile(res.playerProfile);
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
    <div className="flex flex-1 flex-col items-center bg-zinc-50 px-4 py-16 dark:bg-black">
      <div className="flex w-full max-w-3xl flex-col gap-6">
        <Link href="/" className="text-sm text-zinc-500 hover:underline dark:text-zinc-400">
          &larr; back to search
        </Link>

        {loading && <p className="text-zinc-600 dark:text-zinc-400">Loading...</p>}
        {error && <p className="text-red-600 dark:text-red-400">{error}</p>}
        {!loading && !error && !asset && (
          <p className="text-zinc-600 dark:text-zinc-400">Player not found.</p>
        )}

        {asset && (
          <>
            {/* header */}
            <div className="flex items-center gap-6 rounded-xl border border-black/10 bg-white p-8 dark:border-white/15 dark:bg-zinc-900">
              {profile?.imageUrl && (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={profile.imageUrl}
                  alt={asset.displayName}
                  className="h-28 w-28 rounded-full object-cover"
                />
              )}
              <div>
                <h1 className="text-3xl font-semibold tracking-tight text-black dark:text-zinc-50">
                  {asset.displayName}
                </h1>
                {profile?.nameInHomeCountry && profile.nameInHomeCountry !== asset.displayName && (
                  <p className="text-sm text-zinc-500 dark:text-zinc-400">
                    {profile.nameInHomeCountry}
                  </p>
                )}
                <p className="mt-1 text-zinc-600 dark:text-zinc-400">
                  {[
                    asset.symbol,
                    profile?.shirtNumber,
                    profile?.position,
                    profile?.clubName,
                  ]
                    .filter(Boolean)
                    .join(" - ")}
                </p>
                {profile?.citizenship && profile.citizenship.length > 0 && (
                  <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                    {profile.citizenship.join(", ")}
                  </p>
                )}
              </div>
            </div>

            {/* price row */}
            <div className="grid grid-cols-2 gap-4">
              <div className="rounded-xl border border-black/10 bg-white p-6 dark:border-white/15 dark:bg-zinc-900">
                <p className="text-sm text-zinc-500 dark:text-zinc-400">Market value</p>
                <p className="mt-1 font-mono text-2xl text-black dark:text-zinc-50">
                  {profile ? money(profile.marketValue) : "-"}
                </p>
              </div>
              <div className="rounded-xl border border-black/10 bg-white p-6 dark:border-white/15 dark:bg-zinc-900">
                <p className="text-sm text-zinc-500 dark:text-zinc-400">Platform price</p>
                <p className="mt-1 font-mono text-2xl text-black dark:text-zinc-50">
                  ${asset.initialPrice.toFixed(2)}
                </p>
              </div>
            </div>

            {profile?.description && (
              <p className="rounded-xl border border-black/10 bg-white p-6 text-sm text-zinc-700 dark:border-white/15 dark:bg-zinc-900 dark:text-zinc-300">
                {profile.description}
              </p>
            )}

            {/* stat grid */}
            {profile && (
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                <StatBox label="Height" value={profile.height ? `${profile.height} cm` : null} />
                <StatBox label="Foot" value={profile.foot} />
                <StatBox label="Born in" value={[profile.placeOfBirthCity, profile.placeOfBirthCountry].filter(Boolean).join(", ")} />
                <StatBox label="Other positions" value={profile.positionOther.join(", ")} />
                <StatBox label="Club joined" value={profile.clubJoined} />
                <StatBox label="Contract until" value={profile.clubContractExpires} />
                <StatBox label="Agent" value={profile.agentName} />
                <StatBox label="Outfitter" value={profile.outfitter} />
                <StatBox
                  label="transfermarkt id"
                  value={asset.externalId}
                />
              </div>
            )}

            {profile && profile.transfers.length > 0 && (
              <div className="rounded-xl border border-black/10 bg-white p-6 dark:border-white/15 dark:bg-zinc-900">
                <p className="mb-3 text-xs uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
                  Transfer history
                </p>
                <div className="flex flex-col gap-2">
                  {profile.transfers.map((t, i) => (
                    <div
                      key={i}
                      className="flex items-center justify-between border-b border-black/5 pb-2 text-sm last:border-0 dark:border-white/10"
                    >
                      <span className="text-zinc-700 dark:text-zinc-300">
                        {t.clubFromName} &rarr; {t.clubToName}
                      </span>
                      <span className="text-zinc-500 dark:text-zinc-400">
                        {t.season} {t.marketValue > 0 ? `- ${money(t.marketValue)}` : ""}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
