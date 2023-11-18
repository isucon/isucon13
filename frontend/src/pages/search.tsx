import styled from '@emotion/styled';
import Grid from '@mui/joy/Grid';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import React from 'react';
import { useSearchParams } from 'react-router-dom';
import { useLiveStreamsSearch } from '~/api/hooks';
import { VideoThumbnail } from '~/components/video/thumbnail';

export default function SearchResultPage(): React.ReactElement {
  // get query
  const [searchParams] = useSearchParams();
  const tag = searchParams.get('q');

  const liveSterams = useLiveStreamsSearch({
    tag: tag ?? '',
  });

  return (
    <>
      <Stack sx={{ mx: 2, my: 3 }} gap={3}>
        <Container>
          <Typography level="h3">
            検索結果 <i>{tag}</i>
          </Typography>

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
        </Container>
      </Stack>
    </>
  );
}

const Container = styled.div`
  width: 1200px;
  margin: 0 auto;
`;
