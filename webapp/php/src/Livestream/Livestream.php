<?php

declare(strict_types=1);

namespace IsuPipe\Livestream;

use IsuPipe\Top\Tag;
use IsuPipe\User\User;
use JsonSerializable;
use UnexpectedValueException;

class Livestream implements JsonSerializable
{
    /**
     * @param list<Tag> $tags
     * @throws UnexpectedValueException
     */
    public function __construct(
        public int $id,
        public User $owner,
        public string $title,
        public string $description,
        public string $playlistUrl,
        public string $thumbnailUrl,
        public array $tags,
        public int $startAt,
        public int $endAt,
    ) {
        foreach ($tags as $tag) {
            assert($tag instanceof Tag);
        }
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'owner' => $this->owner,
            'title' => $this->title,
            'description' => $this->description,
            'playlist_url' => $this->playlistUrl,
            'thumbnail_url' => $this->thumbnailUrl,
            'tags' => $this->tags,
            'start_at' => $this->startAt,
            'end_at' => $this->endAt,
        ];
    }
}
