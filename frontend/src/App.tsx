import { useEffect, useState } from "react";
import "./App.css";
import { FetchPlaylist, GetSettings } from "../wailsjs/go/main/App";
import { ytdlp } from "../wailsjs/go/models";
import UrlInput from "./components/UrlInput";
import VideoInfo from "./components/VideoInfo";
import FormatPicker from "./components/FormatPicker";
import PlaylistPicker, { PlaylistQuality } from "./components/PlaylistPicker";
import DownloadQueue from "./components/DownloadQueue";
import SettingsPanel from "./components/SettingsPanel";
import { useDownloads } from "./hooks/useDownloads";
import { isLanguage, useI18n } from "./i18n";

// yt-dlp's --flat-playlist entries sometimes carry just a bare video ID in
// `url` rather than a full link (YouTube in particular). Fall back to
// building a watch URL from the entry's id when it isn't already a URL.
function resolveEntryUrl(entry: ytdlp.PlaylistEntry): string {
    if (entry.url && entry.url.startsWith("http")) return entry.url;
    return `https://www.youtube.com/watch?v=${entry.id}`;
}

function App() {
    const { t, setLang } = useI18n();
    const [url, setUrl] = useState("");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [metadata, setMetadata] = useState<ytdlp.Metadata | null>(null);
    const [selectedFormatId, setSelectedFormatId] = useState("");
    const [showSettings, setShowSettings] = useState(false);

    const [selectedEntries, setSelectedEntries] = useState<Set<number>>(new Set());
    const [playlistQuality, setPlaylistQuality] = useState<PlaylistQuality>("best");
    const [playlistSubmitting, setPlaylistSubmitting] = useState(false);

    const { items, startDownload, cancelDownload } = useDownloads();

    useEffect(() => {
        GetSettings().then((s) => {
            document.documentElement.dataset.theme = s.Theme;
            if (isLanguage(s.Language)) setLang(s.Language);
        });
    }, [setLang]);

    async function handleFetch() {
        setLoading(true);
        setError(null);
        setMetadata(null);
        setSelectedFormatId("");
        setSelectedEntries(new Set());
        try {
            // Always fetch flat: it's a no-op for a single video (still
            // returns full formats), and gives a fast entry listing for a
            // playlist without resolving every video's metadata upfront.
            const result = await FetchPlaylist(url);
            setMetadata(result);
            if (result._type === "playlist" && result.entries) {
                setSelectedEntries(new Set(result.entries.map((_, i) => i)));
            }
        } catch (err) {
            setError(String(err));
        } finally {
            setLoading(false);
        }
    }

    async function handleDownload() {
        if (!metadata) return;
        // Proxy/Cookies/OutputDir/OutputTemplate are left unset here so the
        // backend fills them in from the user's saved settings.
        const options = new ytdlp.DownloadOptions({
            URL: url,
            FormatID: selectedFormatId,
        });
        try {
            await startDownload(url, metadata.title, options);
        } catch (err) {
            setError(String(err));
        }
    }

    function toggleEntry(index: number) {
        setSelectedEntries((prev) => {
            const next = new Set(prev);
            if (next.has(index)) next.delete(index);
            else next.add(index);
            return next;
        });
    }

    function toggleAllEntries(checked: boolean) {
        if (!metadata?.entries) return;
        setSelectedEntries(checked ? new Set(metadata.entries.map((_, i) => i)) : new Set());
    }

    async function handleDownloadPlaylist() {
        if (!metadata?.entries) return;
        setPlaylistSubmitting(true);
        try {
            for (const index of Array.from(selectedEntries).sort((a, b) => a - b)) {
                const entry = metadata.entries[index];
                const options = new ytdlp.DownloadOptions({
                    URL: resolveEntryUrl(entry),
                    FormatID: playlistQuality === "worst" ? "worst" : "",
                    AudioOnly: playlistQuality === "audio",
                });
                try {
                    await startDownload(options.URL, entry.title, options);
                } catch (err) {
                    setError(String(err));
                }
            }
        } finally {
            setPlaylistSubmitting(false);
        }
    }

    const isPlaylist = metadata?._type === "playlist";

    if (showSettings) {
        return (
            <div id="App">
                <SettingsPanel onClose={() => setShowSettings(false)} />
            </div>
        );
    }

    return (
        <div id="App">
            <div className="app-header">
                <h1>yt-dlp GUI</h1>
                <button className="settings-link" onClick={() => setShowSettings(true)}>
                    {t("settings.title")}
                </button>
            </div>
            <UrlInput value={url} onChange={setUrl} onSubmit={handleFetch} loading={loading} />

            {error && (
                <div className="error-box">
                    <span>{error}</span>
                    <button
                        className="error-box-dismiss"
                        onClick={() => setError(null)}
                        aria-label={t("error.close")}
                    >
                        ×
                    </button>
                </div>
            )}

            {metadata && !isPlaylist && (
                <>
                    <VideoInfo metadata={metadata} />
                    <FormatPicker
                        formats={metadata.formats ?? []}
                        durationSeconds={metadata.duration}
                        selectedFormatId={selectedFormatId}
                        onSelect={setSelectedFormatId}
                    />
                    <button className="download-btn" onClick={handleDownload}>
                        {t("format.startDownload")}
                    </button>
                </>
            )}

            {metadata && isPlaylist && (
                <PlaylistPicker
                    entries={metadata.entries ?? []}
                    selected={selectedEntries}
                    onToggle={toggleEntry}
                    onToggleAll={toggleAllEntries}
                    quality={playlistQuality}
                    onQualityChange={setPlaylistQuality}
                    onDownload={handleDownloadPlaylist}
                    downloading={playlistSubmitting}
                />
            )}

            <DownloadQueue items={items} onCancel={cancelDownload} />
        </div>
    );
}

export default App;
