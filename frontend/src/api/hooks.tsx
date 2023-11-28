/* eslint-disable @typescript-eslint/explicit-module-boundary-types */

import React from 'react';
import useSWR, { type SWRConfiguration } from 'swr';
import { getThumbnailUrl } from '~/assets';
import { Parameter$get$livestream$search } from './apiClient';
import { HTTPError, apiClient } from './client';

export function useUserMe(config?: SWRConfiguration) {
  return useSWR(
    `/user/me`,
    async () => {
      try {
        return await apiClient.get$user$me({});
      } catch (e) {
        if (e instanceof HTTPError) {
          switch (e.response.status) {
            case 403:
              return null;
            case 401:
              return null;
          }
        }
        throw e;
      }
    },
    config,
  );
}

export function useLiveSelfStreams(config?: SWRConfiguration) {
  return useSWR(`/livestream`, () => apiClient.get$livestream({}), config);
}

export function useLiveStreamsSearch(
  params: Parameter$get$livestream$search | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    params && `/livestream/search?${encodeParam(params)}`,
    () =>
      apiClient.get$livestream$search({
        parameter: params!,
      }),
    config,
  );
}

export function useLiveStream(id: string | null, config?: SWRConfiguration) {
  return useSWR(
    id && `/livestream/${id}/`,
    () =>
      apiClient.get$livestream$_livestreamid({
        parameter: {
          livestreamid: id ?? '',
        },
      }),
    config,
  );
}

export function useLiveUserStream(
  username: string | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    username && `user/${username}/livestream`,
    () =>
      apiClient.get$user$livestream({
        parameter: {
          username: username ?? '',
        },
      }),
    config,
  );
}

export function useLiveStreamComment(
  id: string | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    id && `/livestream/${id}/livecomment`,
    () =>
      apiClient.get$livestream$_livestreamid$livecomment({
        parameter: {
          livestreamid: id ?? '',
          limit: 100,
        },
      }),
    config,
  );
}

export function useLiveStreamReaction(
  id: string | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    id && `/livestream/${id}/reaction`,
    () =>
      apiClient.get$livestream$_livestreamid$reaction({
        parameter: {
          livestreamid: id ?? '',
          limit: 100,
        },
      }),
    config,
  );
}

export function useLiveStreamNgWords(
  id: string | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    id && `/livestream/${id}/ngwords`,
    () =>
      apiClient.get$livecomment$livecommentid$ngwords({
        parameter: {
          livestreamid: id ?? '',
        },
      }),
    config,
  );
}

export function useLiveStreamReports(
  id: string | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    id && `/livestream/${id}/report`,
    () =>
      apiClient.get$livecomment$livecommentid$reports({
        parameter: {
          livestreamid: id ?? '',
        },
      }),
    config,
  );
}

export function useLiveStreamStatistics(
  id: string | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    id && `/livestream/${id}/statistics`,
    () =>
      apiClient.get$livestream$_livestreamid$statistics({
        parameter: {
          livestreamid: id ?? '',
        },
      }),
    config,
  );
}

export function useUser(username: string | null, config?: SWRConfiguration) {
  return useSWR(
    username && `/user/${username}`,
    () =>
      apiClient.get$user$username({
        parameter: {
          username: username ?? '',
        },
      }),
    config,
  );
}

export function useUserStatistics(
  username: string | null,
  config?: SWRConfiguration,
) {
  return useSWR(
    username && `/user/${username}/statistics`,
    () =>
      apiClient.get$user$statistics({
        parameter: {
          username: username ?? '',
        },
      }),
    config,
  );
}

export function useLiveStreamMeasure(id: string | null) {
  React.useEffect(() => {
    if (!id) {
      return;
    }
    apiClient.post$livestream$livestreamid$enter({
      parameter: {
        livestreamid: id,
      },
    });

    return () => {
      apiClient.delete$livestream$livestreamid$exit({
        parameter: {
          livestreamid: id,
        },
      });
    };
  }, [id]);
}

export function useTags(config?: SWRConfiguration) {
  return useSWR('/tags', () => apiClient.get$tag(), config);
}

export interface UseMediaResponse {
  id: number;
  name: string;
  playlist_url: string;
  thumbnail_url: string;
}

export function useMedia(id: string | number, config?: SWRConfiguration) {
  if (import.meta.env.USE_REMOTE_MEDIA) {
    const url = `https://media.xiii.isucon.dev/api/${id}/live/`;
    // eslint-disable-next-line react-hooks/rules-of-hooks
    return useSWR(
      url,
      () => fetch(url).then((res) => res.json() as Promise<UseMediaResponse>),
      config,
    );
  } else {
    const idNum = Number(id) || 0;
    // eslint-disable-next-line react-hooks/rules-of-hooks
    return useSWR(
      `/local_media/${id}`,
      async () =>
        ({
          id: idNum,
          name: `Local Media ${idNum}`,
          playlist_url: `https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8`,
          thumbnail_url: getThumbnailUrl(idNum),
        }) satisfies UseMediaResponse,
      config,
    );
  }
}

function encodeParam(params: Object): string {
  const p = Object.entries(params);
  p.sort(([key1], [key2]) => key1.localeCompare(key2));
  return p.map(([key, value]) => `${key}=${value}`).join('&');
}
