import { act } from 'react';
import { fireEvent, render, waitFor } from '@testing-library/react-native';
import JudgmentScreen from './JudgmentScreen';
import { mockPublicFetch } from './test/mockApi';

function selectAllFactors(getByLabelText: (label: string) => any) {
  return act(async () => {
    fireEvent.press(getByLabelText('Transgression score 3, Good, not selected'));
    fireEvent.press(getByLabelText('Creativity score 2, Fair, not selected'));
    fireEvent.press(getByLabelText('Commitment score 4, Bold, not selected'));
    fireEvent.press(getByLabelText('Presentation score 1, Weak, not selected'));
  });
}

test('JudgmentScreen renders four factor selectors and a disabled review button', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={() => {}}
      onCancel={() => {}}
    />
  );

  expect(getByText('Transgression')).toBeTruthy();
  expect(getByText('Creativity')).toBeTruthy();
  expect(getByText('Commitment')).toBeTruthy();
  expect(getByText('Presentation')).toBeTruthy();

  const reviewButton = getByLabelText('Review Judgment, disabled');
  expect(reviewButton).toBeTruthy();
  expect(getByText('Review Judgment')).toBeTruthy();
});

test('selecting all four scores enables the Review Judgment button', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const { getByLabelText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={() => {}}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await waitFor(() => {
    expect(getByLabelText('Review Judgment, enabled')).toBeTruthy();
  });
});

test('review button opens a Judgment Receipt with selected scores', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={() => {}}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await waitFor(() => {
    expect(getByText('Judgment Receipt')).toBeTruthy();
    expect(getByLabelText('Judgment Receipt factor Transgression score 3')).toBeTruthy();
    expect(getByLabelText('Judgment Receipt factor Creativity score 2')).toBeTruthy();
    expect(getByLabelText('Judgment Receipt factor Commitment score 4')).toBeTruthy();
    expect(getByLabelText('Judgment Receipt factor Presentation score 1')).toBeTruthy();
    expect(getByLabelText('File Judgment')).toBeTruthy();
  });
});

test('filing judgment calls submitJudgment with the selected scores and session token', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL, init?: RequestInit) => {
    const requestUrl = url.toString();
    if (requestUrl.includes('/v1/jumps/jump_1/judgment')) {
      expect(init?.method).toBe('POST');
      const body = JSON.parse(init?.body as string);
      expect(body).toMatchObject({
        commitment: 4,
        transgression: 3,
        creativity: 2,
        presentation: 1,
      });
      expect(init?.headers).toEqual(
        expect.objectContaining({
          Authorization: 'Bearer tok_abc',
        })
      );
      return Response.json({ id: 'judge_1', jumpId: 'jump_1', commitment: 4, transgression: 3, creativity: 2, presentation: 1 });
    }
    return Response.json({});
  });

  const onSubmitted = jest.fn();

  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={onSubmitted}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await waitFor(() => {
    expect(getByText('Judgment Receipt')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByLabelText('File Judgment'));
  });

  await waitFor(() => {
    expect(onSubmitted).toHaveBeenCalledTimes(1);
  });
});

test('renders a clear error state when submission fails', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => {
    return new Response(
      JSON.stringify({ error: 'forbidden', message: 'Author Grace Period is still active.' }),
      { status: 403, headers: { 'Content-Type': 'application/json' } }
    );
  });

  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={() => {}}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await waitFor(() => {
    expect(getByText('Judgment Receipt')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByLabelText('File Judgment'));
  });

  await waitFor(() => {
    expect(getByText(/Author Grace Period is still active/)).toBeTruthy();
  });
});

test('guest judgment: filing calls submitJudgment with guestSessionId and no Authorization header', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL, init?: RequestInit) => {
    const requestUrl = url.toString();
    if (requestUrl.includes('/v1/jumps/jump_1/judgment')) {
      expect(init?.method).toBe('POST');
      const body = JSON.parse(init?.body as string);
      expect(body).toMatchObject({
        guestSessionId: 'guest_session_test123',
        commitment: 4,
        transgression: 3,
        creativity: 2,
        presentation: 1,
      });
      expect(init?.headers).not.toHaveProperty('Authorization');
      return Response.json({ id: 'judge_guest_1', jumpId: 'jump_1', guestSessionId: 'guest_session_test123', commitment: 4, transgression: 3, creativity: 2, presentation: 1 });
    }
    return Response.json({});
  });

  const onSubmitted = jest.fn();

  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      guestSessionId="guest_session_test123"
      onSubmitted={onSubmitted}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await waitFor(() => {
    expect(getByText('Judgment Receipt')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByLabelText('File Judgment'));
  });

  await waitFor(() => {
    expect(onSubmitted).toHaveBeenCalledTimes(1);
  });
});

test('guest cap error shows account-conversion prompt with Sign In button', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => {
    return new Response(
      JSON.stringify({ error: 'guest_cap', message: 'Guest Judgment cap reached.' }),
      { status: 403, headers: { 'Content-Type': 'application/json' } }
    );
  });

  const onSignIn = jest.fn();

  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      guestSessionId="guest_session_capped"
      onSubmitted={() => {}}
      onCancel={() => {}}
      onSignIn={onSignIn}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await waitFor(() => {
    expect(getByText('Judgment Receipt')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByLabelText('File Judgment'));
  });

  await waitFor(() => {
    expect(getByText(/You've reached the guest judging limit/)).toBeTruthy();
    expect(getByText(/Create an account to keep judging/)).toBeTruthy();
    expect(getByLabelText('Sign In to Keep Judging')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByLabelText('Sign In to Keep Judging'));
  });

  expect(onSignIn).toHaveBeenCalledTimes(1);
});

test('factor buttons include named tier labels in accessibility for screen readers', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const { getByLabelText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={() => {}}
      onCancel={() => {}}
    />
  );

  // Each score button should include a tier name, not just a number
  expect(getByLabelText('Transgression score 1, Weak, not selected')).toBeTruthy();
  expect(getByLabelText('Transgression score 2, Fair, not selected')).toBeTruthy();
  expect(getByLabelText('Transgression score 3, Good, not selected')).toBeTruthy();
  expect(getByLabelText('Transgression score 4, Bold, not selected')).toBeTruthy();
});

test('error and cap prompt containers announce changes via accessibilityLiveRegion', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => {
    return new Response(
      JSON.stringify({ error: 'forbidden', message: 'Author Grace Period is still active.' }),
      { status: 403, headers: { 'Content-Type': 'application/json' } }
    );
  });

  const { getByLabelText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={() => {}}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await act(async () => {
    fireEvent.press(getByLabelText('File Judgment'));
  });

  await waitFor(() => {
    const errorContainer = getByLabelText('Judgment submission error');
    expect(errorContainer.props.accessibilityLiveRegion).toBe('polite');
  });
});

test('network failure shows friendly message with Try Again button', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  let callCount = 0;
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => {
    callCount++;
    if (callCount === 1) throw new TypeError('Network request failed');
    return Response.json({ id: 'judge_1' }, { status: 201 });
  });

  const onSubmitted = jest.fn();
  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={onSubmitted}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await act(async () => {
    fireEvent.press(getByLabelText('File Judgment'));
  });

  await waitFor(() => {
    expect(getByText(/Network error/)).toBeTruthy();
    expect(getByLabelText('Try Again')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByLabelText('Try Again'));
  });

  await waitFor(() => {
    expect(onSubmitted).toHaveBeenCalledTimes(1);
  });
});

test('409 already-judged shows clear message that Judgment was already filed', async () => {
  const detail = {
    id: 'jump_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Carried it cleanly',
    runningAverage: 0,
    judgmentCount: 0,
  };

  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => {
    return new Response(
      JSON.stringify({ error: 'already_judged', message: 'Judge has already submitted a Judgment for this Jump.' }),
      { status: 409, headers: { 'Content-Type': 'application/json' } }
    );
  });

  const { getByLabelText, getByText } = await render(
    <JudgmentScreen
      jumpId="jump_1"
      detail={detail}
      session={{ access_token: 'tok_abc', user: { id: 'u1' } }}
      onSubmitted={() => {}}
      onCancel={() => {}}
    />
  );

  await selectAllFactors(getByLabelText);

  await act(async () => {
    fireEvent.press(getByLabelText('Review Judgment, enabled'));
  });

  await act(async () => {
    fireEvent.press(getByLabelText('File Judgment'));
  });

  await waitFor(() => {
    expect(getByText(/You already judged this Jump/)).toBeTruthy();
  });
});
