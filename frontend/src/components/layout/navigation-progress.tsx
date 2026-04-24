import { useNavigation } from 'react-router-dom';

export const NavigationProgress = () => {
  const navigation = useNavigation();
  const isLoading = navigation.state !== 'idle';

  return (
    <div
      className={isLoading ? 'navigation-progress navigation-progress--active' : 'navigation-progress'}
      aria-hidden={!isLoading}
    />
  );
};
