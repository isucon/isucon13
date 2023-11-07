import Avatar from '@mui/joy/Avatar';
import Box from '@mui/joy/Box';
import Button from '@mui/joy/Button';
import Card from '@mui/joy/Card';
import List from '@mui/joy/List';
import ListItem from '@mui/joy/ListItem';
import Sheet from '@mui/joy/Sheet';
import Skeleton from '@mui/joy/Skeleton';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { AiOutlinePlus } from 'react-icons/ai';
import { Link, useParams } from 'react-router-dom';
import { apiClient } from '~/api/client';
import {
  useLiveStream,
  useLiveStreamNgWords,
  useLiveStreamStatistics,
  useMedia,
  useUserStatistics,
} from '~/api/hooks';
import { NewNgWordDialog } from '~/components/console/ngword';
import { useGlobalToastQueue } from '~/components/toast/toast';
import LiveComment from '~/components/video/comment';
import { Video } from '~/components/video/video';

export default function WatchPage(): React.ReactElement {
  const { id } = useParams();
  const liveStream = useLiveStream(id ?? null);
  const idNum = id ? parseInt(id) : null;
  const media = useMedia(id ?? '');
  const ngWords = useLiveStreamNgWords(id ?? null);
  const statistics = useLiveStreamStatistics(id ?? null);
  const userStatistics = useUserStatistics(
    liveStream.data?.owner?.name ?? null,
  );

  const toast = useGlobalToastQueue();
  const [openNgWordDialog, setOpenNgWordDialog] =
    React.useState<boolean>(false);
  const onNgWordSubmit = React.useCallback(
    async (word: string) => {
      if (!id) {
        return;
      }
      await apiClient.post$livestream$livestreamid$moderate({
        parameter: {
          livestreamid: id,
        },
        requestBody: {
          ng_word: word,
        },
      });
      await ngWords.mutate();
      toast.add(
        {
          type: 'success',
          title: 'NGワードを追加しました',
          message: `「${word}」をNGワードに追加しました。`,
        },
        {
          timeout: 3000,
        },
      );
      setOpenNgWordDialog(false);
    },
    [id],
  );

  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={2}>
      <Stack direction="row" gap={2}>
        <Stack direction="column" sx={{ flexBasis: '300px', flexGrow: 1 }}>
          <Box sx={{ mb: 3 }}>
            <Video playlist={media.data?.playlist_url} />
          </Box>

          <Typography level="h3">動画タイトル</Typography>
          <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
            <Link to="/user">
              <Avatar />
            </Link>
            <div>
              {liveStream.data === undefined ? (
                <Skeleton variant="text" level="title-sm" width={100} />
              ) : (
                <Link
                  to={`/user/${liveStream.data?.owner?.name}`}
                  style={{ textDecoration: 'none' }}
                >
                  <Typography level="title-sm">
                    {liveStream.data?.owner?.display_name}
                  </Typography>
                </Link>
              )}
              <Typography level="body-sm" component="div">
                <Stack direction="row" spacing={2}>
                  <span>ランキング {userStatistics.data?.rank}位</span>
                </Stack>
              </Typography>
            </div>
          </Stack>
          <Card variant="plain" sx={{ my: 2 }}>
            <Stack direction="row" columnGap={2} flexWrap="wrap">
              {statistics.data === undefined ? (
                <>
                  <Skeleton variant="text" level="title-sm" width={100} />
                  <Skeleton variant="text" level="title-sm" width={100} />
                  <Skeleton variant="text" level="title-sm" width={100} />
                </>
              ) : (
                <>
                  <Typography level="title-sm">
                    ランキング {statistics.data?.rank}位
                  </Typography>
                  <Typography level="title-sm">
                    {statistics.data?.viewers_count}人が視聴中
                  </Typography>
                  <Typography level="title-sm">
                    最大チップ額 {statistics.data?.max_tip}ISU
                  </Typography>
                  <Typography level="title-sm">
                    {statistics.data?.total_reactions}リアクション
                  </Typography>
                  <Typography level="title-sm">
                    {statistics.data?.total_reports}通報
                  </Typography>
                </>
              )}

              <Typography level="title-sm">2時間前にライブ配信開始</Typography>
            </Stack>
            <Typography
              level="body-md"
              sx={{ whiteSpace: 'pre-wrap' }}
            >{`説明文\n2行目\n三行目`}</Typography>
          </Card>
        </Stack>

        <Stack sx={{ flexBasis: '250px', flexGrow: 1, gap: 0 }}>
          <LiveComment
            type="real"
            livestream_id={idNum ?? 0}
            is_loading={idNum === null || liveStream.isLoading}
          />
        </Stack>

        <NewNgWordDialog
          isOpen={openNgWordDialog}
          onClose={() => setOpenNgWordDialog(false)}
          onSubmit={onNgWordSubmit}
        />
        <Stack direction="column" sx={{ flexBasis: '300px', flexGrow: 1 }}>
          <Stack direction="row" sx={{ mb: 1, alignItems: 'center' }}>
            <Typography level="title-lg">NG Word</Typography>
            <Button
              variant="plain"
              startDecorator={<AiOutlinePlus size="1rem" />}
              sx={{ ml: 'auto' }}
              onClick={() => setOpenNgWordDialog(true)}
            >
              追加
            </Button>
          </Stack>
          <Sheet variant="outlined" sx={{ borderRadius: 'sm' }}>
            <List>
              {ngWords.data?.map((word) => (
                <ListItem key={word.id}>{word.word}</ListItem>
              ))}
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
