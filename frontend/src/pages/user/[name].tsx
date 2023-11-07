import styled from '@emotion/styled';
import AspectRatio from '@mui/joy/AspectRatio';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { useParams } from 'react-router-dom';
import { useUser, useUserStatistics } from '~/api/hooks';
import { iconUrl } from '~/api/icon';
import { VideoThumbnail } from '~/components/video/thumbnail';

export default function UserPage(): React.ReactElement {
  const { name: username } = useParams();
  const user = useUser(username ?? null);
  const userStatistics = useUserStatistics(username ?? null);

  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={3}>
      <Container>
        <Stack direction="row" gap={2} alignItems="center">
          <AspectRatio ratio={1} sx={{ borderRadius: '50%', width: '100px' }}>
            <img src={iconUrl(user.data?.name)} loading="lazy" />
          </AspectRatio>
          <Stack spacing={1}>
            <Typography level="h3">{user.data?.name}</Typography>
            <Stack direction="row" spacing={2}>
              <Typography level="body-sm">
                ランキング{userStatistics.data?.rank}位
              </Typography>
              <Typography level="body-sm">
                コメント数 {userStatistics.data?.total_livecomments}
              </Typography>
              <Typography level="body-sm">
                リアクション数 {userStatistics.data?.total_reactions}
              </Typography>
              <Typography level="body-sm">
                Tip数 {userStatistics.data?.total_tip}ISU
              </Typography>
              <Typography level="body-sm">
                視聴者数 {userStatistics.data?.viewers_count}
              </Typography>
            </Stack>
            <Typography level="body-sm">{user.data?.description}</Typography>
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
                <VideoThumbnail liveSteram={{ id: index }} />
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
