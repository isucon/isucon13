import styled from '@emotion/styled';
import { Typography } from '@mui/joy';
import AspectRatio from '@mui/joy/AspectRatio';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { VideoThumbnail } from '~/components/video/thumbnail';

export default function UserPage(): React.ReactElement {
  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={3}>
      <Cover
        style={{
          backgroundImage:
            'url(https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=1000)',
          backgroundRepeat: 'no-repeat',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
        }}
      />
      <Container>
        <Stack direction="row" gap={2} alignItems="center">
          {/* <Avatar size="lg" /> */}
          <AspectRatio ratio={1} sx={{ borderRadius: '50%', width: '100px' }}>
            <img
              src="https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=400"
              loading="lazy"
            />
          </AspectRatio>
          <Stack spacing={1}>
            <Typography level="h3">ユーザー名</Typography>
            <Stack direction="row" spacing={2}>
              <Typography level="body-sm">登録日: 2021/01/01</Typography>
              <Typography level="body-sm">
                チャンネル登録者数: 1234人
              </Typography>
            </Stack>
            <Typography level="body-sm">説明文</Typography>
          </Stack>
        </Stack>
      </Container>
      <Container>
        <Grid
          container
          spacing={3}
          columns={3}
          flexGrow={1}
          sx={{ padding: 2 }}
        >
          {Array(10)
            .fill(0)
            .map((_, index) => (
              <Grid key={index} xs={1}>
                <VideoThumbnail />
              </Grid>
            ))}
        </Grid>
      </Container>
    </Stack>
  );
}

const Container = styled.div`
  width: 1000px;
  margin: 0 auto;
`;

const Cover = styled.div`
  width: 100%;
  height: 200px;
  background-color: #ccc;
`;
