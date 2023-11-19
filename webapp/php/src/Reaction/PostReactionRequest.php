<?php

declare(strict_types=1);

namespace IsuPipe\Reaction;

use JsonException;
use TypeError;
use UnexpectedValueException;

class PostReactionRequest
{
    public function __construct(
        public string $emojiName,
    ) {
    }

    /**
     * @param string $json
     * @return PostReactionRequest
     * @throws UnexpectedValueException
     */
    public static function fromJson(string $json): PostReactionRequest
    {
        try {
            $data = json_decode($json, flags: JSON_THROW_ON_ERROR);
        } catch (JsonException $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }

        if (!isset($data->emoji_name)) {
            throw new UnexpectedValueException('required fields are missing');
        }

        try {
            return new PostReactionRequest(
                emojiName: $data->emoji_name,
            );
        } catch (TypeError $e) {
            throw new UnexpectedValueException(
                message: $e->getMessage(),
                previous: $e,
            );
        }
    }
}
