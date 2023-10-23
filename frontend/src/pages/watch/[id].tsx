import { Typography } from '@mui/joy';
import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Card from '@mui/joy/Card';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { Link } from 'react-router-dom';
import LiveComment from '~/components/video/comment';

export default function WatchPage(): React.ReactElement {
  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={2}>
      <Stack direction="row" gap={2}>
        <AspectRatio ratio={16 / 9} sx={{ flexBasis: '600px', flexGrow: 3 }}>
          <video />
        </AspectRatio>
        <Stack sx={{ flexBasis: '250px', flexGrow: 1, gap: 0 }}>
          <LiveComment />
        </Stack>
      </Stack>

      <Stack direction="row" gap={2}>
        <Stack sx={{ flexBasis: '600px', flexGrow: 3 }}>
          <Typography level="h3">動画タイトル</Typography>
          <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
            <Link to="/user">
              <Avatar />
            </Link>
            <div>
              <Link to="/user" style={{ textDecoration: 'none' }}>
                <Typography level="title-sm">チャンネル名</Typography>
              </Link>
              <Typography level="body-sm">
                <Stack direction="row" spacing={2}>
                  <span>チャンネル登録者数1234人</span>
                </Stack>
              </Typography>
            </div>
            <div>
              <Button variant="outlined" color="neutral" sx={{ marginLeft: 3 }}>
                チャンネル登録
              </Button>
            </div>
          </Stack>
          <Card variant="plain" sx={{ my: 2 }}>
            <Stack direction="row" spacing={2}>
              <Typography level="title-sm">1,234人が視聴中</Typography>
              <Typography level="title-sm">2時間前にライブ配信開始</Typography>
            </Stack>
            <Typography
              level="body-md"
              sx={{ whiteSpace: 'pre-wrap' }}
            >{`説明文\n2行目\n三行目`}</Typography>
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
