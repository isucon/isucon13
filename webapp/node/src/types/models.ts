export interface UserModel {
  id: number
  name: string
  display_name: string
  password: string
  description: string
}

export interface IconModel {
  id: number
  user_id: number
  image: ArrayBuffer
}

export interface ThemeModel {
  id: number
  user_id: number
  dark_mode: boolean
}

export interface LivestreamsModel {
  id: number
  user_id: number
  title: string
  description: string
  playlist_url: string
  thumbnail_url: string
  start_at: number
  end_at: number
}

export interface ReservationSlotsModel {
  id: number
  slot: number
  start_at: number
  end_at: number
}

export interface TagsModel {
  id: number
  name: string
}

export interface LivestreamTagsModel {
  id: number
  livestream_id: number
  tag_id: number
}

export interface LivestreamViewersHistoryModel {
  id: number
  user_id: number
  livestream_id: number
  created_at: number
}

export interface LivecommentsModel {
  id: number
  user_id: number
  livestream_id: number
  comment: string
  tip: number
  created_at: number
}

export interface LivecommentReportsModel {
  id: number
  user_id: number
  livestream_id: number
  livecomment_id: number
  created_at: number
}

export interface NgWordsModel {
  id: number
  user_id: number
  livestream_id: number
  word: string
  created_at: number
}

export interface ReactionsModel {
  id: number
  user_id: number
  livestream_id: number
  emoji_name: string
  created_at: number
}
