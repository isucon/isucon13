import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Box from '@mui/joy/Box';
import Card from '@mui/joy/Card';
import Divider from '@mui/joy/Divider';
import Skeleton from '@mui/joy/Skeleton';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { Link, useParams } from 'react-router-dom';
import { useLiveStream, useMedia } from '~/api/hooks';
import LiveComment from '~/components/video/comment';
import { Video } from '~/components/video/video';

export default function WatchPage(): React.ReactElement {
  const { id } = useParams();
  const liveStream = useLiveStream(id ?? null);
  const idNum = id ? parseInt(id) : null;
  const media = useMedia(id ?? '');

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
        <Stack sx={{ flexBasis: '600px', flexGrow: 3 }}>
          <Typography level="h3">{liveStream.data?.title}</Typography>
          <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
            <Link to="/user">
              <Avatar />
            </Link>
            <div>
              {liveStream.data === undefined ? (
                <Skeleton variant="text" level="title-sm" width={100} />
              ) : (
                <Link to="/user" style={{ textDecoration: 'none' }}>
                  <Typography level="title-sm">
                    {liveStream.data?.owner?.display_name}
                  </Typography>
                </Link>
              )}
              <Typography level="body-sm">
                <span>チャンネル登録者数****人</span>
              </Typography>
            </div>
            {/* <div>
              <Button variant="outlined" color="neutral" sx={{ marginLeft: 3 }}>
                チャンネル登録
              </Button>
            </div> */}
          </Stack>
          <Card variant="plain" sx={{ my: 2 }}>
            <Stack direction="row" spacing={2}>
              {liveStream.data === undefined ? (
                <Skeleton variant="text" level="title-sm" width={100} />
              ) : (
                <Typography level="title-sm">
                  {liveStream.data?.viewers_count}人が視聴中
                </Typography>
              )}
              <Typography level="title-sm">
                ****時間前にライブ配信開始
              </Typography>
            </Stack>
            <Typography level="body-md" sx={{ whiteSpace: 'pre-wrap' }}>
              {liveStream.data?.description}
            </Typography>
            <Divider sx={{ my: 2, mx: 0 }} />
            <Typography level="body-md" sx={{ whiteSpace: 'pre-wrap' }}>
              {liveStream.data?.owner?.description}
            </Typography>
          </Card>
        </Stack>
        <Stack sx={{ flexBasis: '250px', flexGrow: 1, px: '16px' }}>
          <Typography level="h3">Related Live</Typography>
          <Stack spacing={2} sx={{ my: 1 }}>
            {Array(5)
              .fill(0)
              .map((_, i) => (
                <Link key={i} to="/" style={{ textDecoration: 'none' }}>
                  <Stack direction="row" spacing={2}>
                    <AspectRatio sx={{ borderRadius: 10, flexBasis: '35%' }}>
                      <img
                        src="https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=400"
                        loading="lazy"
                      />
                    </AspectRatio>
                    <Stack sx={{ marginTop: 1 }}>
                      <Typography level="title-md">
                        動画タイトル{i + 1}
                      </Typography>
                      <Typography level="body-sm">チャンネル名</Typography>
                      <Typography level="body-sm">1,234人が視聴中</Typography>
                    </Stack>
                  </Stack>
                </Link>
              ))}
          </Stack>
        </Stack>
      </Stack>
    </Stack>
  );
}
