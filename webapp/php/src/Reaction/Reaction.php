<?php

declare(strict_types=1);

namespace IsuPipe\Reaction;

use IsuPipe\User\User;
use IsuPipe\Livestream\Livestream;
use JsonSerializable;

class Reaction implements JsonSerializable
{
    public function __construct(
        public int $id,
        public string $emojiName,
        public User $user,
        public Livestream $livestream,
        public int $createdAt,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'emoji_name' => $this->emojiName,
            'user' => $this->user,
            'livestream' => $this->livestream,
            'created_at' => $this->createdAt,
        ];
    }
}
