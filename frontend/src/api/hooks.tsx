/* eslint-disable @typescript-eslint/explicit-module-boundary-types */

import useSWR, { type SWRConfiguration } from 'swr';
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
  params: Parameter$get$livestream$search,
  config?: SWRConfiguration,
) {
  return useSWR(
    `/livestream/search?${encodeParam(params)}`,
    () =>
      apiClient.get$livestream$search({
        parameter: params,
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
        },
      }),
    config,
  );
}

export function useTags(config?: SWRConfiguration) {
  return useSWR('/tags', () => apiClient.get$tag(), config);
}

function encodeParam(params: Object): string {
  const p = Object.entries(params);
  p.sort(([key1], [key2]) => key1.localeCompare(key2));
  return p.map(([key, value]) => `${key}=${value}`).join('&');
}
