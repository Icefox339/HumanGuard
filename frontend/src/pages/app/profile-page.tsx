import { ProfileCard } from '@/features/profile/profile-card';
import { ChangePasswordForm } from '@/features/profile/change-password-form';
import { UpdateAvatarForm } from '@/features/profile/update-avatar-form';

export const ProfilePage = () => <section className="space-y-4"><div className="space-y-4"><ProfileCard /><ChangePasswordForm /><UpdateAvatarForm /></div></section>;
