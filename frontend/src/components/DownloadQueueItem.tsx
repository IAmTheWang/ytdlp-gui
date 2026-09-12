import type { DownloadItem } from "../types/download";
import { formatBytes, formatETA, formatSpeed } from "../lib/format";
import { useI18n } from "../i18n";

const CANCELABLE: ReadonlySet<DownloadItem["Status"]> = new Set([
    "queued",
    "downloading",
    "post_processing",
]);

interface DownloadQueueItemProps {
    item: DownloadItem;
    onCancel: (jobId: string) => void;
}

function DownloadQueueItem({ item, onCancel }: DownloadQueueItemProps) {
    const { t } = useI18n();
    const percent = item.Status === "completed" ? 100 : Math.min(100, Math.max(0, item.Percent));

    return (
        <div className="queue-item">
            <div className="queue-item-header">
                <span className="queue-item-title">{item.title || item.url}</span>
                <span className={`queue-item-status status-${item.Status}`}>
                    {t(`queue.status.${item.Status}`)}
                </span>
            </div>

            <div className="progress-bar">
                <div className="progress-bar-fill" style={{ width: `${percent}%` }} />
            </div>

            <div className="queue-item-details">
                {item.Status === "downloading" && (
                    <span>
                        {percent.toFixed(1)}%
                        {item.TotalBytes > 0 && ` · ${formatBytes(item.DownloadedBytes)} / ${formatBytes(item.TotalBytes)}`}
                        {item.Speed > 0 && ` · ${formatSpeed(item.Speed)}`}
                        {item.ETA > 0 && ` · ${t("queue.remaining", { time: formatETA(item.ETA) })}`}
                    </span>
                )}
                {item.Status === "post_processing" && <span>{item.Message}</span>}
                {item.Status === "failed" && <span className="queue-item-error">{item.Message}</span>}

                {CANCELABLE.has(item.Status) && (
                    <button className="cancel-btn" onClick={() => onCancel(item.JobID)}>
                        {t("queue.cancel")}
                    </button>
                )}
            </div>
        </div>
    );
}

export default DownloadQueueItem;
