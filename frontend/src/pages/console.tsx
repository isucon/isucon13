import styled from '@emotion/styled';
import { Typography } from '@mui/joy';
import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Button from '@mui/joy/Button';
import Card from '@mui/joy/Card';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import React from 'react';
import { Link } from 'react-router-dom';
import { Schemas } from '~/api/types';

export default function StreamerConsolePage(): React.ReactElement {
  return (
    <Stack sx={{ mx: 2, my: 3 }} gap={3}>
      <Container>
        <Typography level="h3">配信一覧</Typography>
        <Stack sx={{ display: 'block', my: 3 }}>
          <Button>予約配信を作成</Button>
        </Stack>

        <Grid
          container
          spacing={3}
          columns={1}
          flexGrow={1}
          sx={{ padding: 2 }}
        >
          {Array(10)
            .fill(0)
            .map((_, index) => (
              <Grid key={index} xs={1}>
                <LiveItem
                  liveSteram={{ id: index, title: 'title', user_id: 123 }}
                />
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

interface LiveItemProps {
  liveSteram: Schemas.Livestream;
}
function LiveItem(props: LiveItemProps): React.ReactElement {
  return (
    <Link
      to={`/watch/${props.liveSteram.id}`}
      style={{ textDecoration: 'none' }}
    >
      <Card>
        <Grid container columns={4} gap={3}>
          <Grid xs={1}>
            <AspectRatio sx={{ borderRadius: 10 }}>
              <img
                src="https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=400"
                loading="lazy"
              />
            </AspectRatio>
          </Grid>

          <Grid xs={2}>
            <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
              <Avatar />
              <div>
                <Typography level="title-sm">
                  {props.liveSteram.title}
                </Typography>
                <Typography level="body-sm" component="div">
                  <Stack direction="row" spacing={2}>
                    <span>{props.liveSteram.user_id}</span>
                    <span>1234人視聴・12分前</span>
                  </Stack>
                </Typography>
              </div>
            </Stack>
          </Grid>
        </Grid>
      </Card>
    </Link>
  );
}
