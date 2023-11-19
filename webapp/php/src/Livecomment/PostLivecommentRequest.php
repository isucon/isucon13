<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use JsonException;
use TypeError;
use UnexpectedValueException;

class PostLivecommentRequest
{
    public function __construct(
        public string $comment,
        public int $tip,
    ) {
    }

    /**
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): PostLivecommentRequest
    {
        try {
            $data = json_decode($json, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }

        if (!isset($data->comment, $data->tip)) {
            throw new UnexpectedValueException('required fields are missing');
        }

        try {
            return new PostLivecommentRequest(
                comment: $data->comment,
                tip: $data->tip,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
