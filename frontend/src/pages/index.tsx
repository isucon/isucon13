import { Typography } from '@mui/joy';
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
import { useLiveStreams } from '~/api/hooks';
import { VideoThumbnail } from '~/components/video/thumbnail';

export default function IndexPage(): React.ReactElement {
  const liveSterams = useLiveStreams();
  // TODO Remove
  console.log(liveSterams.data);

  return (
    <div>
      <Stack
        width={230}
        sx={{
          position: 'fixed',
          top: 0,
          bottom: 0,
          paddingTop: '60px',
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
          {/* {liveSterams.data?.slice(0, 30).map((stream, index) => ( */}
          {new Array(30).fill(0).map((stream, index) => (
            <Grid key={index} xs={1}>
              <VideoThumbnail
                liveSteram={{
                  id: index,
                  user_id: 12345,
                  title: 'title',
                }}
              />
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
