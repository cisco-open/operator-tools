package secret

import (
	"context"
	"time"

	"emperror.dev/errors"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretGetter interface {
	Get() (*K8sSecret, error)
}

type K8sSecret struct {
	Token  []byte
	CACert []byte
}

type ReaderSecretGetterOption struct {
	Client                  client.Client
	ServiceAccountName      string
	ServiceAccountNamespace string
}

type readerSecretGetter struct {
	options *ReaderSecretGetterOption
}

func NewReaderSecretGetter(opt *ReaderSecretGetterOption) (SecretGetter, error) {
	if opt == nil {
		return nil, errors.New("reader-secret option should be set for constructor")
	}

	if opt.Client == nil {
		return nil, errors.New("k8s client should be set for reader-secret getter")
	}

	if opt.ServiceAccountNamespace == "" || opt.ServiceAccountName == "" {
		return nil, errors.New("service account name and namespace should be set for reader-secret getter")
	}

	return &readerSecretGetter{options: opt}, nil
}

func (r *readerSecretGetter) Get() (*K8sSecret, error) {
	ctx := context.Background()

	sa, err := r.getReaderSecretServiceAccount(ctx)
	if err != nil {
		return nil, errors.WrapIf(err, "error getting reader secret")
	}

	// After K8s v1.24, Secret objects containing ServiceAccount tokens are no longer auto-generated, so we will have to
	// manually create Secret in order to get the token.
	// Reference: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.24.md#no-really-you-must-read-this-before-you-upgrade
	return r.getOrCreateReaderSecretWithServiceAccount(ctx, sa)
}

func (r *readerSecretGetter) getReaderSecretServiceAccount(ctx context.Context) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}

	saRef := types.NamespacedName{
		Namespace: r.options.ServiceAccountNamespace,
		Name:      r.options.ServiceAccountName,
	}

	err := r.options.Client.Get(ctx, saRef, sa)
	if err != nil {
		return nil, errors.WrapIff(err, "error getting service account object, service account name: %s, service account namespace: %s",
			saRef.Name,
			saRef.Namespace)
	}

	return sa, nil
}

func (r *readerSecretGetter) getOrCreateReaderSecretWithServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) (*K8sSecret, error) {
	secretObj := &corev1.Secret{}

	readerSecretName := sa.Name + "-token"
	if len(sa.Secrets) != 0 {
		readerSecretName = sa.Secrets[0].Name
	}
	secretObjRef := types.NamespacedName{
		Namespace: sa.Namespace,
		Name:      readerSecretName,
	}

	err := r.options.Client.Get(ctx, secretObjRef, secretObj)
	if err != nil && k8sErrors.IsNotFound(err) {
		secretObj = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: secretObjRef.Namespace,
				Name:      secretObjRef.Name,
				Annotations: map[string]string{
					"kubernetes.io/service-account.name": sa.Name,
				},
			},
			Type: "kubernetes.io/service-account-token",
		}

		err = r.options.Client.Create(ctx, secretObj)
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "creating kubernetes secret failed", "namespace",
				secretObjRef.Namespace,
				"secret name",
				secretObjRef.Name)
		}

		// Wait for token-controller to create token for the reader secret
		return r.waitAndGetReaderSecret(ctx, secretObjRef.Namespace, secretObjRef.Name)
	}

	readerSecret := &K8sSecret{
		Token:  secretObj.Data["token"],
		CACert: secretObj.Data["ca.crt"],
	}

	return readerSecret, nil
}

// backoff waiting of the K8s Secret object to be created
var defaultBackoff = wait.Backoff{
	Duration: time.Second * 3,
	Factor:   1,
	Jitter:   0,
	Steps:    3,
}

func (r *readerSecretGetter) waitAndGetReaderSecret(ctx context.Context, secretNamespace string, secretName string) (*K8sSecret, error) {
	var token, caCert []byte

	secretObjRef := types.NamespacedName{
		Namespace: secretNamespace,
		Name:      secretName,
	}

	err := wait.ExponentialBackoff(defaultBackoff, func() (bool, error) {
		tokenSecret := &corev1.Secret{}
		err := r.options.Client.Get(ctx, secretObjRef, tokenSecret)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				return false, nil
			}

			return false, err
		}

		token = tokenSecret.Data["token"]
		caCert = tokenSecret.Data["ca.crt"]

		if token == nil || caCert == nil {
			return false, nil
		}

		return true, nil
	})

	readerSecret := &K8sSecret{
		Token:  token,
		CACert: caCert,
	}

	return readerSecret, errors.WrapIfWithDetails(err, "fail to wait for the token and CA cert to be generated",
		"secret namespace", secretObjRef.Namespace, "secret name", secretObjRef.Name)
}
