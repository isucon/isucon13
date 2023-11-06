import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Card from '@mui/joy/Card';
import List from '@mui/joy/List';
import ListItem from '@mui/joy/ListItem';
import Sheet from '@mui/joy/Sheet';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { AiOutlinePlus } from 'react-icons/ai';
import { Link, useParams } from 'react-router-dom';
import { apiClient } from '~/api/client';
import { useLiveStream, useLiveStreamNgWords } from '~/api/hooks';
import { NewNgWordDialog } from '~/components/console/ngword';
import { useGlobalToastQueue } from '~/components/toast/toast';
import LiveComment from '~/components/video/comment';

export default function WatchPage(): React.ReactElement {
  const { id } = useParams();
  const liveStream = useLiveStream(id ?? null);
  const idNum = id ? parseInt(id) : null;
  const ngWords = useLiveStreamNgWords(id ?? null);

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
