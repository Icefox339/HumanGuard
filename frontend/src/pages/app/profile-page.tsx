import { useState } from 'react';
import { ProfileCard } from '@/features/profile/profile-card';
import { ChangePasswordForm } from '@/features/profile/change-password-form';
import { UpdateAvatarForm } from '@/features/profile/update-avatar-form';

export const ProfilePage = () => {
  const [profileVersion, setProfileVersion] = useState(0);

  return (
    <section className="space-y-4">
      <div className="space-y-4">
        <ProfileCard refreshKey={profileVersion} />
        <ChangePasswordForm />
        <UpdateAvatarForm onUpdated={() => setProfileVersion((prev) => prev + 1)} />
      </div>
    </section>
  );
};
