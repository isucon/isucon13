import { Typography } from '@mui/joy';
import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Card from '@mui/joy/Card';
import List from '@mui/joy/List';
import ListItem from '@mui/joy/ListItem';
import ListItemButton from '@mui/joy/ListItemButton';
import Sheet from '@mui/joy/Sheet';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { AiOutlinePlus } from 'react-icons/ai';
import { Link } from 'react-router-dom';
import LiveComment from '~/components/video/comment';

export default function WatchPage(): React.ReactElement {
  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={2}>
      <Stack direction="row" gap={2}>
        <Stack direction="column" sx={{ flexBasis: '300px', flexGrow: 1 }}>
          <AspectRatio ratio={16 / 9} sx={{ mb: 3 }}>
            <video />
          </AspectRatio>

          <Typography level="h3">動画タイトル</Typography>
          <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
            <Link to="/user">
              <Avatar />
            </Link>
            <div>
              <Link to="/user" style={{ textDecoration: 'none' }}>
                <Typography level="title-sm">チャンネル名</Typography>
              </Link>
              <Typography level="body-sm" component="div">
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

        <Stack sx={{ flexBasis: '250px', flexGrow: 1, gap: 0 }}>
          <LiveComment type="random" livestream_id={1} />
        </Stack>
        <Stack direction="column" sx={{ flexBasis: '300px', flexGrow: 1 }}>
          <Stack direction="row" sx={{ mb: 1, alignItems: 'center' }}>
            <Typography level="title-lg">NG Word</Typography>
            <Button
              variant="plain"
              startDecorator={<AiOutlinePlus size="1rem" />}
              sx={{ ml: 'auto' }}
            >
              追加
            </Button>
          </Stack>
          <Sheet variant="outlined" sx={{ borderRadius: 'sm' }}>
            <List>
              <ListItem>
                <ListItemButton>item</ListItemButton>
              </ListItem>
              <ListItem>
                <ListItemButton>item</ListItemButton>
              </ListItem>
              <ListItem>
                <ListItemButton>item</ListItemButton>
              </ListItem>
            </List>
          </Sheet>
        </Stack>
      </Stack>

      <Stack direction="row" gap={2}>
        <Stack sx={{ flexBasis: '600px', flexGrow: 3 }}></Stack>
      </Stack>
    </Stack>
  );
}
