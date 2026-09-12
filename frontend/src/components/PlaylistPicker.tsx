import { ytdlp } from "../../wailsjs/go/models";
import { formatDuration } from "../lib/format";
import { useI18n } from "../i18n";

export type PlaylistQuality = "best" | "worst" | "audio";

interface PlaylistPickerProps {
    entries: ytdlp.PlaylistEntry[];
    selected: ReadonlySet<number>;
    onToggle: (index: number) => void;
    onToggleAll: (checked: boolean) => void;
    quality: PlaylistQuality;
    onQualityChange: (quality: PlaylistQuality) => void;
    onDownload: () => void;
    downloading: boolean;
}

function PlaylistPicker({
    entries,
    selected,
    onToggle,
    onToggleAll,
    quality,
    onQualityChange,
    onDownload,
    downloading,
}: PlaylistPickerProps) {
    const { t } = useI18n();
    const allSelected = entries.length > 0 && selected.size === entries.length;

    return (
        <div className="playlist-picker">
            <div className="playlist-picker-header">
                <span>{t("playlist.selected", { selected: selected.size, total: entries.length })}</span>
                <button onClick={() => onToggleAll(!allSelected)}>
                    {allSelected ? t("playlist.deselectAll") : t("playlist.selectAll")}
                </button>
            </div>

            <div className="playlist-entries">
                {entries.map((entry, index) => (
                    <label key={`${entry.id}-${index}`} className="playlist-entry">
                        <input
                            type="checkbox"
                            checked={selected.has(index)}
                            onChange={() => onToggle(index)}
                        />
                        <span className="playlist-entry-title">{entry.title || entry.id}</span>
                        <span className="playlist-entry-duration">{formatDuration(entry.duration)}</span>
                    </label>
                ))}
            </div>

            <div className="playlist-format-row">
                <select value={quality} onChange={(e) => onQualityChange(e.target.value as PlaylistQuality)}>
                    <option value="best">{t("playlist.qualityBest")}</option>
                    <option value="worst">{t("playlist.qualityWorst")}</option>
                    <option value="audio">{t("playlist.qualityAudio")}</option>
                </select>
                <button
                    className="download-btn"
                    onClick={onDownload}
                    disabled={downloading || selected.size === 0}
                >
                    {downloading ? t("playlist.adding") : t("playlist.download", { count: selected.size })}
                </button>
            </div>
        </div>
    );
}

export default PlaylistPicker;
