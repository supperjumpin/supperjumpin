import { createClient } from "@supabase/supabase-js";
import { useEffect, useState } from "react";
import { Button, SafeAreaView, StyleSheet, Text, TextInput, View } from "react-native";

import { getMe } from "@supperjumpin/api-client";
import type { MeResponse } from "@supperjumpin/api-client";

const supabase = createClient(
  process.env.EXPO_PUBLIC_SUPABASE_URL ?? "",
  process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY ?? "",
);

const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export default function App() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState("Enter an email to request a Supabase magic link.");
  const [profile, setProfile] = useState<MeResponse | null>(null);

  useEffect(() => {
    supabase.auth.getSession().then(async ({ data }) => {
      if (!data.session) {
        return;
      }

      const me = await getMe({ baseUrl: apiBaseUrl, accessToken: data.session.access_token });
      setProfile(me);
      setStatus("Signed in and connected to Supperjumpin.");
    });
  }, []);

  async function sendMagicLink() {
    const { error } = await supabase.auth.signInWithOtp({ email });
    setStatus(error ? error.message : "Check your email for the Supperjumpin magic link.");
  }

  return (
    <SafeAreaView style={styles.screen}>
      <View style={styles.card}>
        <Text style={styles.title}>Supperjumpin</Text>
        <Text style={styles.body}>{status}</Text>
        <TextInput
          autoCapitalize="none"
          keyboardType="email-address"
          onChangeText={setEmail}
          placeholder="player@example.com"
          style={styles.input}
          value={email}
        />
        <Button onPress={sendMagicLink} title="Send magic link" />
        {profile ? (
          <Text style={styles.body}>
            Account {profile.account.id} is attached to Player {profile.player.id}.
          </Text>
        ) : null}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: "#f7efe2",
    justifyContent: "center",
    padding: 24,
  },
  card: {
    backgroundColor: "#fffaf2",
    borderColor: "#2f241d",
    borderRadius: 24,
    borderWidth: 2,
    gap: 16,
    padding: 24,
  },
  title: {
    color: "#2f241d",
    fontSize: 36,
    fontWeight: "900",
  },
  body: {
    color: "#4d3b31",
    fontSize: 16,
  },
  input: {
    backgroundColor: "white",
    borderColor: "#c1673a",
    borderRadius: 12,
    borderWidth: 1,
    fontSize: 16,
    padding: 12,
  },
});
