# Gluestack UI Component Library

We will use Gluestack UI v2 as the component library for the Expo React Native app rather than raw React Native StyleSheet or an alternative library. The existing prototype StyleSheet was vibecoded with no design input and will be replaced wholesale when the app is built out properly in the frontend architecture spike.

Gluestack UI v2 works within Expo's managed workflow, supporting over-the-air updates without native module compilation. NativeWind was the main alternative considered; it offers more control but requires building a custom component set on top, which adds maintenance overhead before the game model is proven. Tamagui was ruled out for its historically difficult Expo managed workflow integration.
