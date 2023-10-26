import { Typography } from '@mui/joy';
import AspectRatio from '@mui/joy/AspectRatio';
import Avatar from '@mui/joy/Avatar';
import Stack from '@mui/joy/Stack';
import formatDistanceToNow from 'date-fns/formatDistanceToNow';
import { ja } from 'date-fns/locale';
import React from 'react';
import { Link } from 'react-router-dom';
import { Schemas } from '~/api/types';

export interface VideoThumbnailProps {
  liveSteram: Schemas.Livestream;
}
export function VideoThumbnail({
  liveSteram,
}: VideoThumbnailProps): React.ReactElement {
  const date = React.useMemo(
    () =>
      liveSteram.created_at
        ? formatDistanceToNow(liveSteram.created_at * 1000, {
            addSuffix: true,
            locale: ja,
          })
        : 'unkown',
    [liveSteram.created_at],
  );

  return (
    <Link to={`/watch/${liveSteram.id}`} style={{ textDecoration: 'none' }}>
      <AspectRatio sx={{ borderRadius: 10 }}>
        <img
          src="https://images.unsplash.com/photo-1527549993586-dff825b37782?auto=format&fit=crop&w=400"
          loading="lazy"
        />
      </AspectRatio>
      <Stack direction="row" spacing={1} sx={{ marginTop: 1 }}>
        <Avatar />
        <div>
          <Typography level="title-sm">{liveSteram.title}</Typography>
          <Typography level="body-sm" component="div">
            <Stack direction="row" spacing={2}>
              <span>{liveSteram.owner?.name}</span>
              <span>
                {liveSteram.viewers_count}人視聴・{date}
              </span>
            </Stack>
          </Typography>
        </div>
      </Stack>
    </Link>
  );
}
