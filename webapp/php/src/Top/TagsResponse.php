<?php

declare(strict_types=1);

namespace IsuPipe\Top;

use JsonSerializable;

class TagsResponse implements JsonSerializable
{
    /**
     * @param list<Tag> $tags
     */
    public function __construct(
        public array $tags,
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
            'tags' => $this->tags,
        ];
    }
}
