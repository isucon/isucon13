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
        } catch (JsonException) {
            throw new UnexpectedValueException();
        }

        if (!isset($data->comment, $data->tip)) {
            throw new UnexpectedValueException();
        }

        try {
            return new PostLivecommentRequest(
                comment: $data->comment,
                tip: $data->tip,
            );
        } catch (TypeError) {
            throw new UnexpectedValueException();
        }
    }
}
