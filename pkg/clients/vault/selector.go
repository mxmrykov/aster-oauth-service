package vault

import (
	"context"
	"errors"
)

func (v *Vault) GetSecret(ctx context.Context, path, variableName string) (string, error) {
	response, err := v.Client.Read(ctx, path)
	if err != nil {
		return "", err
	}

	data, ok := response.Data["data"].(map[string]interface{})
	if !ok {
		return "", errors.New("cannot assert type from vault response")
	}

	return data[variableName].(string), nil
}

func (v *Vault) GetSecretRepo(ctx context.Context, path string) (map[string]string, error) {
	response, err := v.Client.Read(ctx, path)
	if err != nil {
		return nil, err
	}

	data, ok := response.Data["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("cannot assert type from vault response")
	}

	return func() map[string]string {
		n := make(map[string]string)

		for k, v := range data {
			n[k] = v.(string)
		}

		return n
	}(), nil
}
