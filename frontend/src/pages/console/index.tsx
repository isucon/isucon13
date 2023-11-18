import styled from '@emotion/styled';
import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Card from '@mui/joy/Card';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { Link } from 'react-router-dom';
import { HTTPError, apiClient } from '~/api/client';
import { useLiveSelfStreams } from '~/api/hooks';
import { Schemas } from '~/api/types';
import { NewLiveDialog, NewLiveFormValue } from '~/components/console/newlive';
import { useGlobalToastQueue } from '~/components/toast/toast';

export default function StreamerConsolePage(): React.ReactElement {
  const [open, setOpen] = React.useState<boolean>(false);
  const liveStreams = useLiveSelfStreams();
  const toast = useGlobalToastQueue();
  const onSubmitNewLive = React.useCallback(async (form: NewLiveFormValue) => {
    try {
      await apiClient.post$livestream$reservation({
        requestBody: {
          title: form.title,
          description: form.description,
          tags: form.tags,
          collaborators: [],
          start_at: new Date(form.startAt).valueOf() / 1000,
          end_at: new Date(form.endAt).valueOf() / 1000,
        },
      });
    } catch (e) {
      let message = '配信の作成に失敗しました';
      if (e instanceof HTTPError) {
        if (e.response.status === 400) {
          const body = await e.response.json();
          if (body.message) {
            message = body.message;
          }
        }
      }
      toast.add(
        {
          type: 'error',
          title: '配信の作成に失敗しました',
          message,
        },
        {
          timeout: 5000,
        },
      );
      throw e;
    }
  }, []);

  return (
    <>
      <NewLiveDialog
        isOpen={open}
        onClose={() => setOpen(false)}
        onSubmit={onSubmitNewLive}
      />
      <Stack sx={{ mx: 2, my: 3 }} gap={3}>
        <Container>
          <Typography level="h3">配信一覧</Typography>
          <Stack sx={{ display: 'block', my: 3 }}>
            <Button onClick={() => setOpen(true)}>予約配信を作成</Button>
          </Stack>

          <Grid
            container
            spacing={3}
            columns={1}
            flexGrow={1}
            sx={{ padding: 2 }}
          >
            {/* {Array(10)
              .fill(0) */
            liveStreams.data?.map((stream) => (
              <Grid key={stream.id} xs={1}>
                <LiveItem liveSteram={stream} />
              </Grid>
            ))}
          </Grid>
        </Container>
      </Stack>
    </>
  );
}

const Container = styled.div`
  width: 1000px;
  margin: 0 auto;
`;

interface LiveItemProps {
  liveSteram: Schemas.Livestream;
}
function LiveItem(props: LiveItemProps): React.ReactElement {
  return (
    <Link
      to={`/console/live/${props.liveSteram.id}`}
      style={{ textDecoration: 'none' }}
    >
      <Card>
        <Grid container columns={4} spacing={3}>
          <Grid xs={1}>
            <AspectRatio sx={{ borderRadius: 10 }}>
              <img
                src="https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=400"
                loading="lazy"
              />
            </AspectRatio>
          </Grid>

          <Grid xs={2}>
            <Stack direction="column" spacing={1} sx={{ marginTop: 1 }}>
              <Typography level="title-sm">{props.liveSteram.title}</Typography>
              <Typography level="body-sm" component="div">
                <Stack direction="row" spacing={2}>
                  <span>1234人視聴・12分前</span>
                </Stack>
                <Stack direction="row" spacing={2}>
                  <span>開始 2023-10-10 12:23</span>
                  <span>終了 2023-10-10 12:23</span>
                </Stack>
              </Typography>
            </Stack>
          </Grid>
        </Grid>
      </Card>
    </Link>
  );
}
