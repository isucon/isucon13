import { Typography } from '@mui/joy';
import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Divider from '@mui/joy/Divider';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { BiSolidVideoRecording } from 'react-icons/bi';
import {
  BsFillHouseDoorFill,
  BsCollectionPlay,
  BsClockHistory,
  BsCircleFill,
} from 'react-icons/bs';
import { Link } from 'react-router-dom';

export default function IndexPage(): React.ReactElement {
  return (
    <div>
      <Stack
        width={230}
        sx={{
          position: 'fixed',
          top: 0,
          bottom: 0,
          paddingTop: '60px',
          zIndex: -1,
        }}
      >
        <Stack direction="column" spacing={0} sx={{ padding: 2 }}>
          <SidebarButton startDecorator={<BsFillHouseDoorFill size="20px" />}>
            ホーム
          </SidebarButton>
          <SidebarButton startDecorator={<BsCollectionPlay size="20px" />}>
            登録チャンネル
          </SidebarButton>
          <Divider
            orientation="horizontal"
            sx={{ marginTop: 2, marginBottom: 2 }}
          />
          <SidebarButton startDecorator={<BsClockHistory size="20px" />}>
            再生履歴
          </SidebarButton>
          <SidebarButton startDecorator={<BiSolidVideoRecording size="20px" />}>
            ライブ履歴
          </SidebarButton>
          <Divider
            orientation="horizontal"
            sx={{ marginTop: 2, marginBottom: 2 }}
          />
          <Typography
            level="body-md"
            sx={{ color: 'neutral', paddingLeft: 2, marginBottom: 1 }}
          >
            登録チャンネル
          </Typography>
          {Array(10)
            .fill(0)
            .map((_, i) => (
              <SidebarButton
                key={i}
                startDecorator={<BsCircleFill size="20px" color="#aaa" />}
              >
                {`チャンネル ${i + 1}`}
              </SidebarButton>
            ))}
        </Stack>
      </Stack>
      <Stack
        direction="row"
        spacing={2}
        sx={{
          paddingLeft: '230px',
        }}
      >
        <Grid
          container
          spacing={3}
          columns={4}
          flexGrow={1}
          sx={{ padding: 2 }}
        >
          {Array(10)
            .fill(0)
            .map((_, index) => (
              <Grid key={index} xs={1}>
                <AspectRatio sx={{ borderRadius: 10 }}>
                  <img
                    src="https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=400"
                    loading="lazy"
                  />
                </AspectRatio>
                <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
                  <Avatar />
                  <div>
                    <Typography level="title-sm">ビデオタイトル</Typography>
                    <Typography level="body-sm">
                      <Stack direction="row" spacing={2}>
                        <span>チャンネル名</span>
                        <span>1234人視聴・12分前</span>
                      </Stack>
                    </Typography>
                  </div>
                </Stack>
              </Grid>
            ))}
        </Grid>
      </Stack>
    </div>
  );
}

function SidebarButton(
  props: Parameters<typeof Button>[0],
): React.ReactElement {
  const p = {
    component: Link,
  };
  return (
    <>
      <Button
        {...p}
        variant="plain"
        color="neutral"
        sx={{ paddingLeft: 2, justifyContent: 'start', fontWeight: 'normal' }}
        {...props}
      />
    </>
  );
}
