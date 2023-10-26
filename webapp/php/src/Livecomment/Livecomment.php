<?php

declare(strict_types=1);

namespace IsuPipe\Livecomment;

use IsuPipe\Livestream\Livestream;
use IsuPipe\User\User;
use JsonSerializable;

class Livecomment implements JsonSerializable
{
    public function __construct(
        public int $id,
        public User $user,
        public Livestream $livestream,
        public string $comment,
        public int $tip,
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
            'user' => $this->user,
            'livestream' => $this->livestream,
            'comment' => $this->comment,
            'tip' => $this->tip,
            'created_at' => $this->createdAt,
        ];
    }
}
