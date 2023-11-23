from dataclasses import dataclass


@dataclass
class Theme:
    id: int
    dark_mode: bool


@dataclass
class ThemeModel:
    id: int
    user_id: int
    dark_mode: bool

    def __init__(self, id: int, user_id: int, dark_mode: int) -> None:
        self.id = id
        self.user_id = user_id
        self.dark_mode = dark_mode == 1


@dataclass
class UserModel:
    id: int
    name: str
    display_name: str
    description: str
    password: str  # hashed


@dataclass
class User:
    id: int
    name: str
    display_name: str
    description: str
    theme: Theme
    icon_hash: str


@dataclass
class Tag:
    id: int
    name: str


@dataclass
class Tags:
    tags: list[Tag]


@dataclass
class LiveStreamModel:
    id: int
    user_id: int
    title: str
    description: str
    playlist_url: str
    thumbnail_url: str
    start_at: int
    end_at: int

    def __init__(
        self,
        id: int,
        user_id: int,
        title: str,
        description: str,
        playlist_url: str,
        thumbnail_url: str,
        start_at: int | str,
        end_at: int | str,
    ) -> None:
        self.id = id
        self.user_id = user_id
        self.title = title
        self.description = description
        self.playlist_url = playlist_url
        self.thumbnail_url = thumbnail_url
        self.start_at = int(start_at)
        self.end_at = int(end_at)


@dataclass
class LiveStream:
    id: int
    owner: User
    title: str
    description: str
    playlist_url: str
    thumbnail_url: str
    tags: list[Tag]
    start_at: int
    end_at: int


@dataclass
class LiveStreamTagModel:
    id: int
    livestream_id: int
    tag_id: int


@dataclass
class LiveCommentModel:
    id: int
    user_id: int
    livestream_id: int
    comment: str
    tip: int
    created_at: int


@dataclass
class LiveComment:
    id: int
    user: User
    livestream: LiveStream
    comment: str
    tip: int
    created_at: int


@dataclass
class ReactionModel:
    id: int
    emoji_name: str
    user_id: int
    livestream_id: int
    created_at: int


@dataclass
class Reaction:
    id: int
    emoji_name: str
    user: User
    livestream: LiveStream
    created_at: int


@dataclass
class LiveCommentReportModel:
    id: int
    user_id: int
    livestream_id: int
    livecomment_id: int
    created_at: int


@dataclass
class LiveCommentReport:
    id: int
    reporter: User
    livecomment: LiveComment
    created_at: int


@dataclass
class LiveStreamStatistics:
    rank: int
    viewers_count: int
    total_reactions: int
    total_reports: int
    max_tip: int


@dataclass
class LiveStreamRankingEntry:
    livestream_id: int
    score: int


@dataclass
class UserStatistics:
    rank: int
    viewers_count: int
    total_reactions: int
    total_livecomments: int
    total_tip: int
    favorite_emoji: str


@dataclass
class UserRankingEntry:
    username: str
    score: int


@dataclass
class NGWord:
    id: int
    user_id: int
    livestream_id: int
    word: str
    created_at: int


@dataclass
class ModerateResponse:
    word_id: int


@dataclass
class PaymentResult:
    total_tip: int


@dataclass
class ReservationSlotModel:
    id: int
    slot: int
    start_at: int
    end_at: int

    def __init__(
        self,
        id: int,
        slot: int,
        start_at: int | str,
        end_at: int | str,
    ) -> None:
        self.id = id
        self.slot = slot
        self.start_at = int(start_at)
        self.end_at = int(end_at)
