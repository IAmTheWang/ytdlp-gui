import { ytdlp } from "../../wailsjs/go/models";
import { estimatedSize, formatBytes } from "../lib/format";
import { useI18n } from "../i18n";
import type { TranslationKey } from "../i18n";

function describe(format: ytdlp.Format, t: (key: TranslationKey) => string): string {
    const isAudioOnly = format.vcodec === "none";
    const isVideoOnly = format.acodec === "none";
    if (isAudioOnly) return `${t("format.audio")} · ${format.acodec}`;
    if (isVideoOnly) return `${t("format.videoOnly")} · ${format.resolution || format.format_note}`;
    return format.resolution || format.format_note || "combined";
}

interface FormatPickerProps {
    formats: ytdlp.Format[];
    durationSeconds: number;
    selectedFormatId: string;
    onSelect: (formatId: string) => void;
}

function FormatPicker({ formats, durationSeconds, selectedFormatId, onSelect }: FormatPickerProps) {
    const { t } = useI18n();

    // Storyboard/thumbnail-sprite entries (mhtml) report neither a video nor
    // an audio codec; they aren't a real downloadable stream, so they'd
    // otherwise get mislabeled as "audio" by describe() below.
    const downloadable = formats.filter((f) => f.vcodec !== "none" || f.acodec !== "none");
    if (downloadable.length === 0) return null;

    return (
        <div className="format-picker">
            {downloadable.map((f) => {
                const { bytes, exact } = estimatedSize(f, durationSeconds);
                const size = formatBytes(bytes);
                return (
                    <label key={f.format_id} className="format-row">
                        <input
                            type="radio"
                            name="format"
                            value={f.format_id}
                            checked={selectedFormatId === f.format_id}
                            onChange={() => onSelect(f.format_id)}
                        />
                        <span className="format-desc">{describe(f, t)}</span>
                        <span className="format-ext">{f.ext}</span>
                        <span className="format-size">
                            {size ? `${exact ? "" : "~"}${size}` : ""}
                        </span>
                    </label>
                );
            })}
        </div>
    );
}

export default FormatPicker;
