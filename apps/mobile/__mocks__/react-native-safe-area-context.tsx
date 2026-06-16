import React from 'react';
import { View } from 'react-native';

const MockSafeAreaProvider = ({ children }: { children: React.ReactNode }) => children as React.ReactElement;

const MockSafeAreaView = View;

export const SafeAreaProvider = MockSafeAreaProvider;
export const SafeAreaView = MockSafeAreaView;
