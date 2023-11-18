import {
  ApiClient as ApiClientGenerated,
  Client,
  QueryParameters,
  RequestArgs,
} from '~/api/apiClient';

export class HTTPError extends Error {
  constructor(public response: Response) {
    super(`HTTP Error: ${response.status} ${response.statusText}`);

    if ((Error as any).captureStackTrace) {
      (Error as any).captureStackTrace(this, this.constructor);
    }

    Object.setPrototypeOf(this, new.target.prototype);
  }
}

export interface RequestOption {
  timeout?: number;
}

export function decodeParams(
  params: QueryParameters | undefined,
): string | undefined {
  if (!params) {
    return;
  }

  const param = new URLSearchParams();
  for (const key in params) {
    const value = params[key].value;
    if (value === undefined || value === null) {
      continue; // ignore undefined/null field
    }
    param.set(key, value);
  }
  return param.toString();
}

const apiClientImpl: ApiClientGenerated<RequestOption> = {
  request: async (requestArgs: RequestArgs): Promise<any> => {
    const query = decodeParams(requestArgs.queryParameters);
    const origin = apiOrigin(window.location.origin);
    const requestUrl = `${origin}${
      query ? requestArgs.url + '?' + query : requestArgs.url
    }`;
    const response = await fetch(requestUrl, {
      body: JSON.stringify(requestArgs.requestBody),
      headers: {
        ...requestArgs.headers,
      },
      method: requestArgs.httpMethod,
    });

    if (!response.ok) {
      throw new HTTPError(response);
    }

    if (response.headers.get('Content-Type')?.includes('application/json')) {
      return await response.json();
    } else {
      return response.text();
    }
  },
};

export function apiOrigin(origin: string): string {
  if (origin.includes('localhost')) {
    return origin;
  }
  return origin;
}

export const apiClient = new Client<RequestOption>(apiClientImpl, '/api/');
export type ApiClient = typeof apiClient;
