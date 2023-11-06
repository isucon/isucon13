import Box from '@mui/joy/Box';
import React from 'react';
import videojs from 'video.js';
import 'video.js/dist/video-js.css';

export interface VideoProps {
  playlist?: string;
}
export function Video(props: VideoProps): React.ReactElement {
  const containerRef = React.useRef<HTMLDivElement>(null);
  const playerRef = React.useRef<ReturnType<typeof videojs> | null>(null);

  React.useEffect(() => {
    if (!containerRef.current) {
      return;
    }

    const videoElement = document.createElement('video-js');
    containerRef.current.appendChild(videoElement);
    console.log('videoElement', videoElement, containerRef.current);

    const player = videojs(videoElement, {
      autoplay: true,
      controls: true,
      fluid: true,
      sources: [
        {
          src: props.playlist,
          type: 'application/x-mpegURL',
        },
      ],
    });
    playerRef.current = player;

    return () => {
      player.dispose();
      playerRef.current = null;
      videoElement.remove();
    };
  }, [props.playlist]);

  return <Box ref={containerRef} sx={{ width: '100%' }} />;
}
