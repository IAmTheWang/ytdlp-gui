import { ytdlp } from "../../wailsjs/go/models";
import { formatDuration } from "../lib/format";

interface VideoInfoProps {
    metadata: ytdlp.Metadata;
}

function VideoInfo({ metadata }: VideoInfoProps) {
    return (
        <div className="video-info">
            {metadata.thumbnail && (
                <img className="video-thumb" src={metadata.thumbnail} alt="" />
            )}
            <div className="video-meta">
                <div className="video-title">{metadata.title}</div>
                <div className="video-duration">{formatDuration(metadata.duration)}</div>
            </div>
        </div>
    );
}

export default VideoInfo;
