import Avatar from '@mui/joy/Avatar';
import Card from '@mui/joy/Card';
import Divider from '@mui/joy/Divider';
import Skeleton from '@mui/joy/Skeleton';
import Stack from '@mui/joy/Stack';
import Typography from '@mui/joy/Typography';
import { formatDistanceToNow } from 'date-fns';
import { ja } from 'date-fns/locale';
import React from 'react';
import { Link } from 'react-router-dom';
import {
  useLiveStream,
  useLiveStreamMeasure,
  useLiveStreamStatistics,
  useUserStatistics,
} from '~/api/hooks';
import { iconUrl } from '~/api/icon';
import { normalizeUrl } from '~/api/url';

export interface VideoAboutProps {
  id: string | null;
}
export function VideoAbout({ id }: VideoAboutProps): React.ReactElement {
  const liveStream = useLiveStream(id);
  const statistics = useLiveStreamStatistics(id);
  const userStatistics = useUserStatistics(
    liveStream.data?.owner?.name ?? null,
  );
  useLiveStreamMeasure(id);

  const date = React.useMemo(
    () =>
      liveStream.data?.start_at
        ? formatDistanceToNow(liveStream.data.start_at * 1000, {
            addSuffix: true,
            locale: ja,
          })
        : 'unkown',
    [liveStream.data?.start_at],
  );

  return (
    <Stack sx={{ flexBasis: '600px', flexGrow: 3 }}>
      <Typography level="h3">{liveStream.data?.title}</Typography>
      <Stack direction="row" columnGap={2} flexWrap="wrap">
        <Link to={normalizeUrl(`/user`, liveStream.data?.owner?.name)}>
          <Avatar src={iconUrl(liveStream.data?.owner?.name)} />
        </Link>
        <div>
          {liveStream.data === undefined ? (
            <Skeleton variant="text" level="title-sm" width={100} />
          ) : (
            <Link
              to={normalizeUrl(`/user`, liveStream.data?.owner?.name)}
              style={{ textDecoration: 'none' }}
            >
              <Typography level="title-sm">
                {liveStream.data?.owner?.display_name}
              </Typography>
            </Link>
          )}
          <Typography level="body-sm">
            <span>ランキング {userStatistics.data?.rank}位</span>
          </Typography>
        </div>
      </Stack>
      <Card variant="plain" sx={{ my: 2 }}>
        <Stack
          direction="row"
          sx={{ maxWidth: '400px', flexWrap: 'wrap' }}
          columnGap={2}
        >
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
          <Typography level="title-sm">{date}にライブ配信開始</Typography>
        </Stack>
        <Typography level="body-md" sx={{ whiteSpace: 'pre-wrap' }}>
          {liveStream.data?.description}
        </Typography>
        <Divider sx={{ my: 2, mx: 0 }} />
        <Typography level="body-md" sx={{ whiteSpace: 'pre-wrap' }}>
          {liveStream.data?.owner?.description}
        </Typography>
      </Card>
    </Stack>
  );
}
