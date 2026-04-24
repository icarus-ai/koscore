package ecdh

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"math/big"
)

var DEBUG = false

// big.IsEven pub.Y.Bit(0) == 0
var (
	k_zero = big.NewInt(0)
	k_one  = big.NewInt(1)
	k_two  = big.NewInt(2)
)

type Provider struct {
	curve  *EllipticCurve
	secret *big.Int
	public EllipticPoint
}

func NewEcdhProvider(curve *EllipticCurve) (ec *Provider, e error) {
	ec = &Provider{curve: curve}
	if ec.secret, e = ec.createSecret(); e != nil {
		return nil, e
	}
	if ec.public, e = ec.createPublic(); e != nil {
		return nil, e
	}
	return ec, nil
}

func NewEcdhProviderSecret(curve *EllipticCurve, secret_key []byte) (ec *Provider, e error) {
	ec = &Provider{curve: curve}
	if ec.secret, e = ec.unpackSecret(secret_key); e != nil {
		return nil, e
	}
	if ec.public, e = ec.createPublic(); e != nil {
		return nil, e
	}
	return ec, nil
}

// 密钥交换
func (ctx *Provider) KeyExchange(pub_key []byte, is_hash bool) ([]byte, error) {
	ss, e := ctx.unpackPublic(pub_key)
	if e != nil {
		return nil, e
	}
	shared, e := ctx.createShared(ctx.secret, ss)
	if e != nil {
		return nil, e
	}
	return ctx.packShared(shared, is_hash), nil
}

// 打包公钥
func (ctx *Provider) PackPublic(compress bool) []byte {
	curve := ctx.curve
	pub := ctx.public
	if compress {
		res := make([]byte, curve.Size+1)
		if (pub.Y.Bit(0) == 0) != (pub.Y.Sign() < 0) {
			// ??? if pub.Y.Bit(0) == 0 {
			res[0] = 0x02
		} else {
			res[0] = 0x03
		}
		xB := ToFixedBytes(pub.X, curve.Size)
		copy(res[1:], xB)
		return res
	}
	res := make([]byte, curve.Size*2+1)
	res[0] = 0x04
	xB := ToFixedBytes(pub.X, curve.Size)
	yB := ToFixedBytes(pub.Y, curve.Size)
	copy(res[1:], xB)
	copy(res[1+curve.Size:], yB)
	return res
}

// 打包私钥
func (ctx *Provider) PackSecret() []byte {
	raw := ctx.secret.Bytes()
	size := len(raw)
	res := make([]byte, size+4)
	copy(res[4:], raw)
	res[3] = byte(size)
	return res[:size+4]
}

func (ctx *Provider) packShared(shared EllipticPoint, is_hash bool) []byte {
	x := ToFixedBytes(shared.X, ctx.curve.Size)
	if !is_hash {
		return x
	}
	hash := md5.Sum(x[:ctx.curve.PackSize])
	return hash[:]
}

func (ctx *Provider) unpackPublic(pub_key []byte) (EllipticPoint, error) {
	curve := ctx.curve
	size := curve.Size
	l := len(pub_key)
	if l != size*2+1 && l != size+1 {
		return EllipticPoint{}, errors.New("length does not match")
	}
	if pub_key[0] == 0x04 {
		x := new(big.Int).SetBytes(pub_key[1 : size+1])
		y := new(big.Int).SetBytes(pub_key[size+1:])
		return NewEllipticPoint(x, y), nil
	}
	px := new(big.Int).SetBytes(pub_key[1:])
	x3 := new(big.Int).Mul(px, px)
	x3.Mul(x3, px)
	ax := new(big.Int).Mul(px, curve.A)
	right := new(big.Int).Add(x3, ax)
	right.Add(right, curve.B)
	right = Mod(right, curve.P)

	// tmp = (P + 1) >> 2
	tmp := new(big.Int).Add(curve.P, k_one)
	tmp.Rsh(tmp, 2)
	py := new(big.Int).Exp(right, tmp, curve.P)

	even := py.Bit(0) == 0
	//if !(even && pub_key[0] == 0x02  || !even && pub_key[0] == 0x03) { py.Sub(curve.P, py) }
	if (!even || pub_key[0] != 0x02) && (even || pub_key[0] != 0x03) {
		py.Sub(curve.P, py)
	}
	return NewEllipticPoint(px, py), nil
}

func (ctx *Provider) unpackSecret(sec []byte) (*big.Int, error) {
	if len(sec) == int(sec[3])+4 {
		return new(big.Int).SetBytes(sec[4:]), nil
	}
	return nil, errors.New("length does not match")
}

func (ctx *Provider) createPublic() (EllipticPoint, error) {
	return ctx.createShared(ctx.secret, ctx.curve.G)
}

func (ctx *Provider) createSecret() (*big.Int, error) {
	size := ctx.curve.Size
	buf := make([]byte, size)

	if DEBUG {
		buf[0], buf[len(buf)-1] = 1, 1
		return new(big.Int).SetBytes(buf), nil
	}

	for { // 安全生成 0 < k < N
		_, e := rand.Read(buf)
		if e != nil {
			return nil, e
		}
		res := new(big.Int).SetBytes(buf)
		// 强制把最高位清零 避免数值永远大于 N
		if len(buf) > 0 {
			buf[0] &= 0x7F
		}
		if res.Sign() > 0 && res.Cmp(ctx.curve.N) < 0 {
			return res, nil
		}
	}
}

func (ctx *Provider) createShared(sec *big.Int, pub EllipticPoint) (point EllipticPoint, err error) {
	curve := ctx.curve
	// sec % curve.N == 0
	if Mod(sec, curve.N).Cmp(k_zero) == 0 || pub.IsDefault() {
		return
	}
	if sec.Sign() < 0 {
		return ctx.createShared(new(big.Int).Neg(sec), pub)
	}

	if !curve.CheckOn(pub) {
		err = errors.New("public key is not on the curve")
		return
	}

	pr := NewJacobianPoint(big.NewInt(0), big.NewInt(1), big.NewInt(0))
	pa := JacobianFromAffine(pub)
	ps := new(big.Int).Set(sec)

	for ps.Sign() > 0 {
		if ps.Bit(0) == 1 {
			pr = ctx.jacobianAdd(pr, pa)
		}
		pa = ctx.jacobianDouble(pa)
		ps.Rsh(ps, 1)
	}
	return ctx.jacobianToAffine(pr), nil
}

func (ctx *Provider) jacobianDouble(p JacobianPoint) JacobianPoint {
	if p.IsInfinity() {
		return p
	}
	curve := ctx.curve
	p2 := curve.P
	x, y, z := p.X, p.Y, p.Z

	// 完全严格按照 C# 原版公式
	yy := Mod(new(big.Int).Mul(y, y), p2)
	s := Mod(new(big.Int).Mul(big.NewInt(4), new(big.Int).Mul(x, yy)), p2)
	z2 := new(big.Int).Mul(z, z)
	z4 := new(big.Int).Mul(z2, z2)
	m := Mod(new(big.Int).Add(new(big.Int).Mul(big.NewInt(3), new(big.Int).Mul(x, x)), new(big.Int).Mul(curve.A, z4)), p2)
	x3 := Mod(new(big.Int).Sub(new(big.Int).Mul(m, m), new(big.Int).Mul(big.NewInt(2), s)), p2)
	y3 := Mod(new(big.Int).Sub(new(big.Int).Mul(m, new(big.Int).Sub(s, x3)), new(big.Int).Mul(big.NewInt(8), new(big.Int).Mul(yy, yy))), p2)
	z3 := Mod(new(big.Int).Mul(big.NewInt(2), new(big.Int).Mul(y, z)), p2)

	return NewJacobianPoint(x3, y3, z3)
}

func (ctx *Provider) jacobianAdd(p1, p2 JacobianPoint) JacobianPoint {
	if p1.IsInfinity() {
		return p2
	}
	if p2.IsInfinity() {
		return p1
	}
	curve := ctx.curve
	p := curve.P

	z1z1 := Mod(new(big.Int).Mul(p1.Z, p1.Z), p)
	z2z2 := Mod(new(big.Int).Mul(p2.Z, p2.Z), p)
	u1 := Mod(new(big.Int).Mul(p1.X, z2z2), p)
	u2 := Mod(new(big.Int).Mul(p2.X, z1z1), p)
	s1 := Mod(new(big.Int).Mul(p1.Y, new(big.Int).Mul(p2.Z, z2z2)), p)
	s2 := Mod(new(big.Int).Mul(p2.Y, new(big.Int).Mul(p1.Z, z1z1)), p)

	if u1.Cmp(u2) == 0 {
		if s1.Cmp(s2) == 0 {
			return ctx.jacobianDouble(p1)
		}
		return NewJacobianPoint(big.NewInt(0), big.NewInt(1), big.NewInt(0))
	}

	h := Mod(new(big.Int).Sub(u2, u1), p)
	hh := Mod(new(big.Int).Mul(h, h), p)
	hhh := Mod(new(big.Int).Mul(h, hh), p)
	r := Mod(new(big.Int).Sub(s2, s1), p)
	v := Mod(new(big.Int).Mul(u1, hh), p)

	// x3 = (r ^ 2 - hhh - 2 * v) % p
	x3 := Mod(new(big.Int).Sub(new(big.Int).Mul(r, r), new(big.Int).Add(hhh, new(big.Int).Mul(k_two, v))), p)
	y3 := Mod(new(big.Int).Sub(new(big.Int).Mul(r, new(big.Int).Sub(v, x3)), new(big.Int).Mul(s1, hhh)), p)
	z3 := Mod(new(big.Int).Mul(new(big.Int).Mul(p1.Z, p2.Z), h), p)
	return NewJacobianPoint(x3, y3, z3)
}

func (ctx *Provider) jacobianToAffine(p JacobianPoint) EllipticPoint {
	if p.IsInfinity() {
		return EllipticPoint{}
	}
	curve := ctx.curve
	zInv, _ := ModInverse(p.Z, curve.P)
	zInv2 := Mod(new(big.Int).Mul(zInv, zInv), curve.P)
	zInv3 := Mod(new(big.Int).Mul(zInv2, zInv), curve.P)
	return NewEllipticPoint(Mod(new(big.Int).Mul(p.X, zInv2), curve.P), Mod(new(big.Int).Mul(p.Y, zInv3), curve.P))
}

// 取模运算 保证结果为正
func Mod(a, b *big.Int) *big.Int {
	m := new(big.Int).Mod(a, b)
	if m.Sign() < 0 {
		m.Add(m, b)
	}
	return m
}

// 扩展欧几里得算法求模逆元
func ModInverse(a, p *big.Int) (*big.Int, error) {
	aa := Mod(a, p)
	t0, t1 := big.NewInt(0), big.NewInt(1)
	r0, r1 := new(big.Int).Set(p), new(big.Int).Set(aa)

	// !r1.is_zero
	for r1.Cmp(k_zero) != 0 {
		quotient := new(big.Int).Div(r0, r1)
		t0, t1 = t1, new(big.Int).Sub(t0, new(big.Int).Mul(quotient, t1))
		r0, r1 = r1, new(big.Int).Sub(r0, new(big.Int).Mul(quotient, r1))
	}
	// r0 > 1 ==> cmp(a,b) | a>b 1 | a==b 0 | a<b -1
	if r0.Cmp(k_one) > 0 {
		return nil, errors.New("inverse does not exist")
	}
	if t0.Sign() < 0 {
		t0.Add(t0, p)
	}
	return t0, nil
}

// 固定长度字节转换
func ToFixedBytes(value *big.Int, size int) []byte {
	b := value.Bytes()
	b_size := len(b)
	if b_size == size {
		return b
	}
	res := make([]byte, size)
	if b_size > size {
		copy(res, b[b_size-size:])
	} else {
		copy(res[size-b_size:], b)
	}
	return res
}

// ***** 椭圆曲线参数 *****

type EllipticCurve struct {
	P, A, B  *big.Int
	G        EllipticPoint
	N, H     *big.Int
	Size     int
	PackSize int
}

// (point.Y ^ 2 - point.X ^ 3 - A * point.X - B) % P == 0
func (c *EllipticCurve) CheckOn(point EllipticPoint) bool {
	if point.IsDefault() {
		return true
	}
	rhs := new(big.Int).Mul(point.X, point.X)
	rhs.Mul(rhs, point.X)
	rhs.Add(rhs, new(big.Int).Mul(c.A, point.X))
	rhs.Add(rhs, c.B)
	diff := new(big.Int).Mul(point.Y, point.Y)
	diff.Sub(diff, rhs)
	return Mod(diff, c.P).Cmp(k_zero) == 0 // diff % p == 0
}

// ***** 椭圆曲线点 *****

type EllipticPoint struct{ X, Y *big.Int }

func NewEllipticPoint(x, y *big.Int) EllipticPoint {
	return EllipticPoint{X: new(big.Int).Set(x), Y: new(big.Int).Set(y)}
}

func (e EllipticPoint) IsDefault() bool { return e.X.Sign() == 0 && e.Y.Sign() == 0 }

// ***** 雅可比点 *****

type JacobianPoint struct{ X, Y, Z *big.Int }

func NewJacobianPoint(x, y, z *big.Int) JacobianPoint {
	return JacobianPoint{
		X: new(big.Int).Set(x),
		Y: new(big.Int).Set(y),
		Z: new(big.Int).Set(z),
	}
}

func (j JacobianPoint) IsInfinity() bool { return j.Z.Sign() == 0 }

func JacobianFromAffine(p EllipticPoint) JacobianPoint {
	if p.IsDefault() {
		return NewJacobianPoint(big.NewInt(0), big.NewInt(1), big.NewInt(0))
	}
	return NewJacobianPoint(p.X, p.Y, big.NewInt(1))
}

// 预定义曲线
var (
	Prime256V1 = initPrime256V1()
	Secp192k1  = initSecp192k1()
	Secp224R1  = initSecp224R1()
)

// ***** 曲线初始化 *****

func initSecp192k1() EllipticCurve {
	// 官方标准 secp192k1 参数
	P, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFEE37", 16)
	A := big.NewInt(0)
	B := big.NewInt(3)
	Gx, _ := new(big.Int).SetString("DB4FF10EC057E9AE26B07D0280B7F4341DA5D1B1EAE06C7D", 16)
	Gy, _ := new(big.Int).SetString("9B2F2F6D9C5628A7844163D015BE86344082AA88D95E2F9D", 16)
	N, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFE26F2FC170F69466A74DEFD8D", 16)
	H := big.NewInt(1)

	return EllipticCurve{
		P: P, A: A, B: B,
		G: NewEllipticPoint(Gx, Gy),
		N: N, H: H,
		Size:     24,
		PackSize: 24,
	}
}

func initPrime256V1() EllipticCurve {
	p, _ := new(big.Int).SetString("FFFFFFFF00000001000000000000000000000000FFFFFFFFFFFFFFFFFFFFFFFF", 16)
	a, _ := new(big.Int).SetString("FFFFFFFF00000001000000000000000000000000FFFFFFFFFFFFFFFFFFFFFFFC", 16)
	b, _ := new(big.Int).SetString("5AC635D8AA3A93E7B3EBBD55769886BC651D06B0CC53B0F63BCE3C3E27D2604B", 16)
	gx, _ := new(big.Int).SetString("6B17D1F2E12C4247F8BCE6E563A440F277037D812DEB33A0F4A13945D898C296", 16)
	gy, _ := new(big.Int).SetString("4FE342E2FE1A7F9B8EE7EB4A7C0F9E162BCE33576B315ECECBB6406837BF51F5", 16)
	n, _ := new(big.Int).SetString("FFFFFFFF00000000FFFFFFFFFFFFFFFFBCE6FAADA7179E84F3B9CAC2FC632551", 16)
	h := big.NewInt(1)
	return EllipticCurve{
		P: p, A: a, B: b,
		G: NewEllipticPoint(gx, gy),
		N: n, H: h, Size: 32, PackSize: 16,
	}
}

func initSecp224R1() EllipticCurve {
	/*
		p , _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000000000000000001", 16)
		a , _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF0000000000000000FFFFFFFE", 16)
		b , _ := new(big.Int).SetString("B4050A850C04B3ABF54132565044B0B7D7BFD8BA270B39432355FFB4", 16)
		gx, _ := new(big.Int).SetString("B70E0CBD6BB4BF7F321390B94A03C1D356C21122343280D6115C1D21", 16)
		gy, _ := new(big.Int).SetString("BD376388B5F723FB4C22DFE6CD4375A05A07476444D5819985007E34", 16)
		n , _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFF16A2E0B8F03E13DD29455C5C2A3D163CBF056D", 16)
	*/
	P, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000000000000000001", 16)
	A, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFE", 16)
	B, _ := new(big.Int).SetString("B4050A850C04B3ABF54132565044B0B7D7BFD8BA270B39432355FFB4", 16)
	Gx, _ := new(big.Int).SetString("B70E0CBD6BB4BF7F321390B94A03C1D356C21122343280D6115C1D21", 16)
	Gy, _ := new(big.Int).SetString("BD376388B5F723FB4C22DFE6CD4375A05A07476444D5819985007E34", 16)
	N, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFF16A2E0B8F03E13DD29455C5C2A3D", 16)
	return EllipticCurve{
		P: P, A: A, B: B,
		G: NewEllipticPoint(Gx, Gy),
		N: N, H: big.NewInt(1),
		Size: 28, PackSize: 16,
	}
}
