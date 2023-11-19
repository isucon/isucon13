<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use JsonException;
use TypeError;
use UnexpectedValueException;

class ReserveLivestreamRequest
{
    /**
     * @param list<int> $tags
     * @throws TypeError
     * @throws UnexpectedValueException
     */
    public function __construct(
        public array $tags,
        public string $title,
        public string $description,
        public string $playlistUrl,
        public string $thumbnailUrl,
        public int $startAt,
        public int $endAt,
    ) {
        foreach ($tags as $tag) {
            if (!is_int($tag)) {
                throw new UnexpectedValueException('parameter $tags must be list of int');
            }
        }
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): ReserveLivestreamRequest
    {
        try {
            $data = json_decode($json, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }

        if (
            !isset(
                $data->tags,
                $data->title,
                $data->description,
                $data->playlist_url,
                $data->thumbnail_url,
                $data->start_at,
                $data->end_at,
            )
        ) {
            throw new UnexpectedValueException('required fields are missing');
        }

        try {
            return new ReserveLivestreamRequest(
                tags: $data->tags,
                title: $data->title,
                description: $data->description,
                playlistUrl: $data->playlist_url,
                thumbnailUrl: $data->thumbnail_url,
                startAt: $data->start_at,
                endAt: $data->end_at,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
