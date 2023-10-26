<?php

declare(strict_types=1);

namespace IsuPipe\Stats;

/**
 * @phpstan-type LivestreamRanking list<LivestreamRankingEntry>
 */
class LivestreamRankingEntry
{
    public function __construct(
        public int $livestreamId,
        public string $title,
        public int $score,
    ) {
    }

    public function compare(LivestreamRankingEntry $other): int
    {
        if ($this->score === $other->score) {
            return $this->title <=> $other->title;
        }

        return $this->score <=> $other->score;
    }
}
