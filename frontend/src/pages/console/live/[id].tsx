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
import { useParams } from 'react-router-dom';
import { apiClient } from '~/api/client';
import {
  useLiveStream,
  useLiveStreamNgWords,
  useLiveStreamReports,
  useMedia,
} from '~/api/hooks';
import { NewNgWordDialog } from '~/components/console/ngword';
import { useGlobalToastQueue } from '~/components/toast/toast';
import { VideoAbout } from '~/components/video/about';
import LiveComment from '~/components/video/comment';
import { Video } from '~/components/video/video';

export default function WatchPage(): React.ReactElement {
  const { id } = useParams();
  const liveStream = useLiveStream(id ?? null);
  const idNum = id ? parseInt(id) : null;
  const media = useMedia(id ?? '');
  const ngWords = useLiveStreamNgWords(id ?? null);
  const reports = useLiveStreamReports(id ?? null, {
    refreshInterval: 3000,
  });

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

          <VideoAbout id={id ?? null} />
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

          <Stack direction="row" sx={{ mt: 3, mb: 1, alignItems: 'center' }}>
            <Typography level="title-lg">通報されたコメント</Typography>
          </Stack>
          <Sheet variant="outlined" sx={{ borderRadius: 'sm' }}>
            <List>
              {reports.data?.map((report) => (
                <ListItem key={report.id}>
                  {report.livecomment?.comment}
                </ListItem>
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
