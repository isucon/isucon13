import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Divider from '@mui/joy/Divider';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import {
  BsFillHouseDoorFill,
  BsCollectionPlay,
  BsFillPersonFill,
} from 'react-icons/bs';
import { MdManageHistory } from 'react-icons/md';
import { Link } from 'react-router-dom';
import { useLiveStreamsSearch } from '~/api/hooks';
import { VideoThumbnail } from '~/components/video/thumbnail';

export default function IndexPage(): React.ReactElement {
  const liveSterams = useLiveStreamsSearch({
    limit: 100,
  });

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
          <SidebarButton startDecorator={<BsFillPersonFill size="20px" />}>
            プロフィール
          </SidebarButton>
          <SidebarButton
            startDecorator={<MdManageHistory size="20px" />}
            {...{ to: '/console' }}
          >
            管理画面
          </SidebarButton>
          <Divider
            orientation="horizontal"
            sx={{ marginTop: 2, marginBottom: 2 }}
          />
          <Typography
            level="body-md"
            sx={{ color: 'neutral', paddingLeft: 2, marginBottom: 1 }}
          >
            配信中
          </Typography>
          {liveSterams.data?.slice(0, 10).map((live, i) => (
            <SidebarButton
              key={live.id ?? i}
              startDecorator={
                <Avatar
                  src={`/api/user/${live.owner?.name ?? ''}/icon`}
                  sx={{ width: '25px', height: '25px' }}
                />
              }
            >
              {live.owner?.display_name}
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
          {liveSterams.data?.map((stream, index) => (
            <Grid key={index} xs={1}>
              <VideoThumbnail liveSteram={stream} />
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
