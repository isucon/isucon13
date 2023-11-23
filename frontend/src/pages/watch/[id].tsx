import Box from '@mui/joy/Box';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { useParams } from 'react-router-dom';
import {
  useLiveStream,
  useLiveStreamMeasure,
  useLiveStreamsSearch,
  useMedia,
} from '~/api/hooks';
import { VideoAbout } from '~/components/video/about';
import LiveComment from '~/components/video/comment';
import { VideoThumbnail } from '~/components/video/thumbnail';
import { Video } from '~/components/video/video';

export default function WatchPage(): React.ReactElement {
  const { id } = useParams();
  const liveStream = useLiveStream(id ?? null);
  const idNum = id ? parseInt(id) : null;
  const media = useMedia(id ?? '');
  useLiveStreamMeasure(id ?? null);

  const firstTag = liveStream.data?.tags?.[0];
  const related = useLiveStreamsSearch(
    firstTag ? { tag: firstTag.name } : null,
  );

  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={2}>
      <Stack direction="row" gap={2}>
        <Box sx={{ flexBasis: '600px', flexGrow: 3 }}>
          <Video playlist={media.data?.playlist_url} />
        </Box>
        <Stack sx={{ flexBasis: '250px', flexGrow: 1, gap: 0 }}>
          <LiveComment
            type="real"
            livestream_id={idNum ?? 0}
            is_loading={idNum === null || liveStream.isLoading}
          />
        </Stack>
      </Stack>

      <Stack direction="row" gap={2}>
        <VideoAbout id={id ?? null} />
        <Stack sx={{ flexBasis: '250px', flexGrow: 1, px: '16px' }}>
          <Typography level="h3">Related Live</Typography>
          <Stack spacing={2} sx={{ my: 1 }}>
            {related.data?.map((livestream) => (
              <VideoThumbnail
                key={livestream.id}
                liveSteram={livestream}
                landscape
              />
            ))}
          </Stack>
        </Stack>
      </Stack>
    </Stack>
  );
}
