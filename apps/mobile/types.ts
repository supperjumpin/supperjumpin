import type {
  JumpCard as JumpCardType,
  JumpDetail as JumpDetailType,
  JumpTombstone as JumpTombstoneType,
  ViewerContext as ViewerContextType,
  PublicFeedResponse,
} from "@supperjumpin/api-client";

export type {
  JumpCardType,
  JumpDetailType,
  JumpTombstoneType,
  ViewerContextType,
  PublicFeedResponse,
};

export interface FeedState {
  jumps: JumpCardType[];
  nextCursor: string | null;
  loading: boolean;
  refreshing: boolean;
  loadingMore: boolean;
  error: string | null;
}
