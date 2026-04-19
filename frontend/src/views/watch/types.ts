import type { ChatBadgeTag } from '../../lib/chatBadges';

export type ViewerSortMode =
  | 'present_new'
  | 'present_old'
  | 'login_az'
  | 'login_za'
  | 'account_new'
  | 'account_old'
  | 'message_high'
  | 'message_low';

export type ChatLine = {
  key: string;
  user: string;
  message: string;
  keyword: boolean;
  userMarked: boolean;
  userIsSus: boolean;
  susTitle?: string;
  fromSent: boolean;
  firstMessage: boolean;
  at: number;
  badgeTags: ChatBadgeTag[];
  createdAtIso?: string;
  chatterUserId?: number | null;
};

export type ChatRow =
  | { kind: 'gap'; key: string; label: string }
  | { kind: 'msg'; line: ChatLine };
