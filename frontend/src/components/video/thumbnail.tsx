import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import formatDistanceToNow from 'date-fns/formatDistanceToNow';
import { ja } from 'date-fns/locale';
import React from 'react';
import { Link } from 'react-router-dom';
import { useLiveStreamStatistics, useMedia } from '~/api/hooks';
import { iconUrl } from '~/api/icon';
import { Schemas } from '~/api/types';
import { normalizeUrl } from '~/api/url';

export interface VideoThumbnailProps {
  liveSteram: Schemas.Livestream;
  landscape?: boolean;
}
export function VideoThumbnail({
  liveSteram,
  landscape,
}: VideoThumbnailProps): React.ReactElement {
  const date = React.useMemo(
    () =>
      liveSteram.start_at
        ? formatDistanceToNow(liveSteram.start_at * 1000, {
            addSuffix: true,
            locale: ja,
          })
        : 'unkown',
    [liveSteram.start_at],
  );
  const media = useMedia(liveSteram.id ?? '');
  const statistics = useLiveStreamStatistics(liveSteram.id?.toString() ?? null);

  return (
    <Link
      to={normalizeUrl(`/watch/${liveSteram.id}`, liveSteram.owner?.name)}
      style={{ textDecoration: 'none' }}
    >
      {landscape ? (
        <Stack direction="row" spacing={2}>
          <AspectRatio sx={{ borderRadius: 10, flexBasis: '35%' }}>
            <img src={media.data?.thumbnail_url} loading="lazy" />
          </AspectRatio>
          <Stack sx={{ marginTop: 1 }}>
            <Typography level="title-md">{liveSteram.title}</Typography>
            <Typography level="body-sm">{liveSteram.owner?.name}</Typography>
            <Typography level="body-sm">
              {statistics.data?.viewers_count}人視聴・{date}
            </Typography>
          </Stack>
        </Stack>
      ) : (
        <>
          <AspectRatio sx={{ borderRadius: 10 }}>
            <img src={media.data?.thumbnail_url} loading="lazy" />
          </AspectRatio>
          <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
            <Avatar src={iconUrl(liveSteram.owner?.name)} />
            <div>
              <Typography level="title-sm">{liveSteram.title}</Typography>
              <Typography level="body-sm" component="div">
                <Stack direction="row" spacing={2}>
                  <span>{liveSteram.owner?.name}</span>
                  <span>
                    {statistics.data?.viewers_count}人視聴・{date}
                  </span>
                </Stack>
              </Typography>
            </div>
          </Stack>
        </>
      )}
    </Link>
  );
}
