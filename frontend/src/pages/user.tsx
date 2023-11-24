import styled from '@emotion/styled';
import AspectRatio from '@mui/joy/AspectRatio';
import Box from '@mui/joy/Box';
import Button from '@mui/joy/Button';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { apiClient } from '~/api/client';
import {
  useLiveUserStream,
  useUser,
  useUserMe,
  useUserStatistics,
} from '~/api/hooks';
import { iconUrl } from '~/api/icon';
import { ChangeIconDialog } from '~/components/account/iconmodal';
import { VideoThumbnail } from '~/components/video/thumbnail';

export default function UserPage(): React.ReactElement {
  // const { name: username } = useParams();
  const username = window.location.hostname.split('.')[0];
  const user = useUser(username ?? null);
  const me = useUserMe();
  const userStatistics = useUserStatistics(username ?? null);
  const isSelf = me.data !== undefined && me.data?.name === username;
  const [isModalOpen, setIsModalOpen] = React.useState<boolean>(false);
  const onIconSubmit = React.useCallback(async (iconBase64: string) => {
    await apiClient.post$icon({
      requestBody: {
        image: iconBase64,
      },
    });
    location.reload();
  }, []);
  const livestreams = useLiveUserStream(username ?? null);

  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={3}>
      <Container>
        <ChangeIconDialog
          isOpen={isModalOpen}
          onSubmit={onIconSubmit}
          onClose={() => setIsModalOpen(false)}
        />
        <Stack direction="row" gap={2} alignItems="center">
          <Box>
            <AspectRatio ratio={1} sx={{ borderRadius: '50%', width: '100px' }}>
              <img src={iconUrl(user.data?.name)} loading="lazy" />
            </AspectRatio>
            {isSelf && (
              <Button
                onClick={() => setIsModalOpen(true)}
                sx={{ width: '100%', mt: 1 }}
                size="sm"
                variant="soft"
              >
                画像変更
              </Button>
            )}
          </Box>
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
              <Typography level="body-sm">
                好きな絵文字
                {userStatistics.data?.favorite_emoji && (
                  <em-emoji
                    shortcodes={`:${userStatistics.data.favorite_emoji}:`}
                    set="twitter"
                  ></em-emoji>
                )}
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
          {livestreams.data?.map((livestream, index) => (
            <Grid key={index} xs={1}>
              <VideoThumbnail liveSteram={livestream} />
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
