// Small display-formatting helpers shared by the metadata/format UI.
// These intentionally mirror the fallback logic in the Go
// internal/ytdlp.Format.EstimatedSize method, since the frontend only gets
// the raw filesize/filesize_approx/tbr fields over the wire.

export function formatDuration(totalSeconds: number): string {
    if (!totalSeconds || totalSeconds <= 0) return "--:--";
    const seconds = Math.round(totalSeconds);
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    const pad = (n: number) => n.toString().padStart(2, "0");
    return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`;
}

export function formatSpeed(bytesPerSecond?: number): string {
    if (!bytesPerSecond || bytesPerSecond <= 0) return "";
    return `${formatBytes(bytesPerSecond)}/s`;
}

export function formatETA(seconds?: number): string {
    if (seconds === undefined || seconds === null || seconds < 0) return "";
    return formatDuration(seconds);
}

export function formatBytes(bytes?: number): string {
    if (!bytes || bytes <= 0) return "";
    const units = ["B", "KB", "MB", "GB"];
    let value = bytes;
    let unitIndex = 0;
    while (value >= 1024 && unitIndex < units.length - 1) {
        value /= 1024;
        unitIndex++;
    }
    return `${value.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

// estimatedSize returns the best available size estimate (bytes, exact) for
// a format, given the parent video's duration in seconds.
export function estimatedSize(
    format: { filesize?: number; filesize_approx?: number; tbr?: number },
    durationSeconds: number
): { bytes: number; exact: boolean } {
    if (format.filesize && format.filesize > 0) {
        return { bytes: format.filesize, exact: true };
    }
    if (format.filesize_approx && format.filesize_approx > 0) {
        return { bytes: format.filesize_approx, exact: false };
    }
    if (format.tbr && format.tbr > 0 && durationSeconds > 0) {
        return { bytes: (format.tbr * durationSeconds * 1000) / 8, exact: false };
    }
    return { bytes: 0, exact: false };
}
